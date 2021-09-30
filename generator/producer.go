package main

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"time"
)

var (
	interval       time.Duration
	kafkBrokers    []string
	producer       sarama.AsyncProducer
	kafkaTopic     string
	config         *sarama.Config
	)

func Initialize() error {
	log.Println("initializing kafka producer")
	interval = time.Second*30
	kafkBrokers = []string{"kafka-service:30999"}
	kafkaTopic = "events"
	config = sarama.NewConfig()
	err := setupKafkaProducer()
	if err != nil {
		return err
	}
	return nil
}

func setupKafkaProducer() error {
	log.Println("setting up kafka producer")
	var err error
	producer, err = sarama.NewAsyncProducer(kafkBrokers, config)
	if err != nil {
		log.Printf("error %s during creation of kafka producer %s", err)
		return err
	}
	log.Println("Kafka producer setup successfully")
	return nil
}

func produceKafkaMessages() {
	//produce kafka messages at intervals
	//if msg size more than 1 MB,  put into object store if available
	for {
		event := Event{
			Id:    uuid.New().String(),
			Name:  "random",
			Dept:  "random",
			EmpId: rand.Int(),
			Time:  time.Now(),
		}
		log.Printf("event %s", event.Id)
		jsonMsg,err := json.Marshal(event)
		if err != nil {
			log.Printf("error during converting event to json %s", err)
			continue
		}
		log.Printf("sending event to kafka")
		sendKafkaMsg(event.Id, jsonMsg)
		time.Sleep(interval)
	}

}

func sendKafkaMsg(eventId string, msg []byte) {
	kMsg := &sarama.ProducerMessage{
		Topic:     kafkaTopic,
		Key:       sarama.StringEncoder(eventId),
		Value:     sarama.ByteEncoder(msg),
	}
	log.Println("formed kafka msg , will send the event now")
	producer.Input() <- kMsg
}

type Event struct {
	Id    string       `json:"id"`
	Name  string       `json:"name"`
	Dept  string       `json:"dept"`
	EmpId int          `json:"empid"`
	Time   time.Time   `json:"time"`
}
