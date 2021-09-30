package main

import "log"

func main(){
	err := Initialize()
	if err != nil {
		log.Println("Unable to start generator due to %s", err)
		panic(err)
	}
	log.Println("setup complete")
	log.Println("starting the generator")
	produceKafkaMessages()
	select {}
}