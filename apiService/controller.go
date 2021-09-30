package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

//Controller ...
type Controller struct {
}


/* Middleware handler to handle all requests for authentication */
func AuthenticationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorizationHeader := req.Header.Get("authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
				token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("There was an error")
					}
					return []byte("secret"), nil
				})
				if error != nil {
					json.NewEncoder(w).Encode(Exception{Message: error.Error()})
					return
				}
				if token.Valid {
					log.Println("TOKEN WAS VALID")
					context.Set(req, "decoded", token.Claims)
					next(w, req)
				} else {
					json.NewEncoder(w).Encode(Exception{Message: "Invalid authorization token"})
				}
			}
		} else {
			json.NewEncoder(w).Encode(Exception{Message: "An authorization header is required"})
		}
	})
}

// Get Authentication token GET /
func (c *Controller) GetSessionToken(w http.ResponseWriter, req *http.Request) {
	var user User
	_ = json.NewDecoder(req.Body).Decode(&user)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"password": user.Password,
	})

	log.Println("Request received to get session key from the device client ", req.RemoteAddr)

	tokenString, error := token.SignedString([]byte("secret"))
	if error != nil {
		fmt.Println(error)
	}

	log.Println("The server is returning session token :",  tokenString)

	json.NewEncoder(w).Encode(JwtToken{Token: tokenString})
}


func (c *Controller) ActivateDevice(w http.ResponseWriter, r *http.Request) {
	var activation Activation
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Println("Error: Unable to read the body of the request.", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := r.Body.Close(); err != nil {
		log.Printf("Error: Unexpected error had occurred. %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("Request received to activate the following device from the device client ", r.RemoteAddr)
	log.Println("Device details specified in the request body: ", string(body))

	serverConfig := config.GetServerConfiguration()
	//fmt.Println(serverConfig)

	signPrivateKeyFile  := serverConfig.KeyStoreLocation + serverConfig.SignPrivateKeyFile
	encryptPublicKeyFile := serverConfig.KeyStoreLocation+ serverConfig.EncryptPublicKeyFile

	if err := json.Unmarshal(body, &activation); err != nil { // unmarshall body contents as a type Candidate
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Println("Error UpdateProduct unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	deviceId := activation.DeviceId

	if deviceId == "" {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		log.Println("Error: Device id is not specified in the request body.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	activatedDevicesList, err := GetActivatedDevicesList()
	if err := r.Body.Close(); err != nil {
		log.Printf("Error: Unable to get activated devices list. %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("Registered storage device Ids: ", activatedDevicesList)

	if isDeviceActivated(deviceId, activatedDevicesList) == false{
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		log.Println("Error: The specified device is not registered and activated.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("The specified device is registered.")
	log.Println("Fetching the SAAS instance access details...")

	saasDetails := ciphers.GetSaasDetails(deviceId)

	log.Println("Generating the digital signature for the SAAS instance access details.")
	payload := ciphers.GetPayload(saasDetails, signPrivateKeyFile, encryptPublicKeyFile)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	log.Println("Returning the following activation payload.")
	log.Println(string(payload))

	_,_ = w.Write(payload)
	return

}

func (c *Controller) FetchCert(w http.ResponseWriter, r *http.Request) {

	log.Println("Fetching cert to connect to cloud.")

	serverConfig := config.GetServerConfiguration()
	cloudPublicKeyFile := serverConfig.KeyStoreLocation + serverConfig.CloudPublicKeyFile

	log.Println("Generating the digital signature for the SAAS instance access details.")

	//TODO: Generate and send certificate based on array serial number.
	byteValue, err := ioutil.ReadFile(cloudPublicKeyFile)
	if err != nil {
		log.Printf("Error: Unable to read certificate file. %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	payload := ciphers.GetPlainPayload(byteValue)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	log.Println("Returning the following certificate payload.")
	log.Println(string(payload))

	_,_ = w.Write(payload)
	return

}

func isDeviceActivated(deviceId string, activatedDeviceList []string) bool {
	for _, activatedDeviceId := range activatedDeviceList {
		if deviceId == activatedDeviceId {
			return true
		}
	}
	return false
}

// Get Events List - GET /
func (c *Controller) GetEvents(w http.ResponseWriter, req *http.Request) {
	//get events from DB
	eventsList, err := GetEventsFromDB()
	if err != nil {
		log.Printf("Error: Unable to get events list. %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	devBytes,err := json.Marshal(eventsList)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(devBytes)

	return
}

func (c *Controller) AddDeviceToList(w http.ResponseWriter, r *http.Request) {
	log.Print("here")
	var activation Activation
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Println("Error: Unable to read the body of the request.", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := r.Body.Close(); err != nil {
		log.Printf("Error: Unexpected error had occurred. %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("Request received to activate the following device from the device client ", r.RemoteAddr)
	log.Println("Device details specified in the request body: ", string(body))

	if err := json.Unmarshal(body, &activation); err != nil { // unmarshall body contents as a type Candidate
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Println("Error UpdateProduct unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	deviceId := activation.DeviceId

	if deviceId == "" {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		log.Println("Error: Device id is not specified in the request body.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	activatedDevicesList, err := GetActivatedDevicesList()
	if err := r.Body.Close(); err != nil {
		log.Printf("Error: Unable to get activated devices list. %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, b := range activatedDevicesList {
		if b == deviceId {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			log.Println("Error: The specified device is already registered and activated.")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("The specified device is already registered and activated."))
			return
		}
	}
	activatedDevicesList,err = AddDevice(deviceId)
	if err != nil {
		log.Print("Unable to activate device due to ",err)
	}
	log.Println("Registered storage device Ids: ", activatedDevicesList)

	log.Println("The specified device is registered.")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	return

}
