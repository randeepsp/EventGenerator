package main

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"strings"
	"time"
)

var (
	kafkBrokers     []string
	consumer        sarama.Consumer
	kafkaTopics     []string
	config          *sarama.Config
	groupConsumerId string
)

/*
we can use consumer group to consume as a group
 */
func Initialize(ctx context.Context) error{

	kafkBrokers = []string{"localhost:9092"}

	config = sarama.NewConfig()
	config.ClientID = "go-kafka-event-consumer"
	config.Consumer.Return.Errors = true

	err := setupKafkaConsumer()
	if err != nil {
		return err
	}
	return nil
}

func setupKafkaConsumer() error {
	var err error
	consumer, err = sarama.NewConsumer(kafkBrokers, config)
	if err != nil {
		log.Printf("error %s during creation of kafka consumer %s", err)
		return err
	}
	kafkaTopics,err = consumer.Topics()
	if err != nil {
		log.Printf("no topics found to consumer error %s", err)
		return err
	}
	return nil
}

func consumeKafkaMessages(ctx context.Context) {
	consumer, errors := consume(kafkaTopics, consumer)
	for {
		select {
		case msg := <-consumer:
			log.Printf("msg received, push to db")
			pushMsgtoDB(msg)
		case consumerError := <-errors:
			log.Printf("Received consumerError ", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
		}
	}

}


func consume(topics []string, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)
	for _, topic := range topics {
		if strings.Contains(topic, "__consumer_offsets") {
			continue
		}
		partitions, _ := master.Partitions(topic)
		// considering only one topic and one partition
		consumer, err := master.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
		if nil != err {
			log.Printf("Topic %v  has Partitions: %v", topic, partitions)
			panic(err)
		}
		log.Printf(" will start consuming topic ", topic)
		go func(topic string, consumer sarama.PartitionConsumer) {
			for {
				select {
				case consumerError := <-consumer.Errors():
					errors <- consumerError
					fmt.Println("consumerError: ", consumerError.Err)

				case msg := <-consumer.Messages():
					consumers <- msg
					fmt.Println("Got message on topic ", topic, msg.Value)
				}
			}
		}(topic, consumer)
	}

	return consumers, errors
}

type Event struct {
	Id    string
	Name  string
	Dept  string
	EmpId int
	Time   time.Time
}
