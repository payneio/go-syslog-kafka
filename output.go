package main

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/mcuadros/go-syslog"
	"log"
	"time"
)

// A KafkaProducer encapsulates a connection to a Kafka cluster.
type KafkaProducer struct {
}

// Returns an initialized KafkaProducer.
func NewKafkaProducer(msgChan syslog.LogPartsChannel, brokers []string, topic string, bufferTime, bufferBytes int) (*KafkaProducer, error) {
	self := &KafkaProducer{}

	clientConfig := sarama.NewClientConfig()
	client, err := sarama.NewClient("gocollector", brokers, clientConfig)
	if err != nil {
		log.Printf("failed to create kafka client", err)
		return nil, err
	}

	producerConfig := sarama.NewProducerConfig()
	producerConfig.Partitioner = sarama.NewRandomPartitioner()
	producerConfig.MaxBufferedBytes = uint32(bufferBytes)
	producerConfig.MaxBufferTime = time.Duration(bufferTime) * time.Millisecond
	producer, err := sarama.NewProducer(client, producerConfig)
	if err != nil {
		log.Printf("failed to create kafka producer", err)
		return nil, err
	}

	go func() {
		for messageParts := range msgChan {
			messageBytes, err := json.Marshal(messageParts)
			if err != nil {
				log.Printf("Could not marshal logParts", err)
			} else {
				producer.QueueMessage(topic, nil, sarama.StringEncoder(messageBytes))
			}
		}
	}()

	log.Printf("kafka producer created")
	return self, nil
}
