package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

const serviceName = "dp-search-data-extractor"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	// Get Config
	config, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error getting config", err)
		os.Exit(1)
	}

	// Create Kafka Producer
	pChannels := kafka.CreateProducerChannels()
	kafkaProducer, err := kafka.NewProducer(ctx, config.KafkaAddr, config.ContentPublishedTopic, pChannels, &kafka.ProducerConfig{
		KafkaVersion: &config.KafkaVersion,
	})
	if err != nil {
		log.Fatal(ctx, "fatal error trying to create kafka producer", err, log.Data{"topic": config.ContentPublishedTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.Channels().LogErrors(ctx, "kafka producer")

	time.Sleep(500 * time.Millisecond)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		e := scanEvent(scanner)
		log.Info(ctx, "sending content-published event", log.Data{"contentPublishedEvent": e})

		bytes, err := schema.ContentPublishedEvent.Marshal(e)
		if err != nil {
			log.Fatal(ctx, "content-published event error", err)
			os.Exit(1)
		}

		// Send bytes to Output channel, after calling Initialise just in case it is not initialised.
		kafkaProducer.Initialise(ctx)
		kafkaProducer.Channels().Output <- bytes
	}
}

// scanEvent creates a ContentPublished event according to the user input
func scanEvent(scanner *bufio.Scanner) *event.ContentPublished {
	fmt.Println("--- [Send Kafka ContentPublished] ---")

	fmt.Println("Please type the URL")
	fmt.Printf("$ ")
	scanner.Scan()
	name := scanner.Text()

	return &event.ContentPublished{
		URL: name,
	}
}
