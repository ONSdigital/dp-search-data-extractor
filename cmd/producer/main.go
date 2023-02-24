package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

const serviceName = "dp-search-data-extractor"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	// Get Config
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error getting config", err)
		os.Exit(1)
	}

	// Create Kafka Producer
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:     cfg.Kafka.Addr,
		Topic:           cfg.Kafka.ContentUpdatedTopic,
		KafkaVersion:    &cfg.Kafka.Version,
		MaxMessageBytes: &cfg.Kafka.MaxBytes,
	}
	if cfg.Kafka.SecProtocol == config.KafkaTLSProtocolFlag {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.Kafka.SecCACerts,
			cfg.Kafka.SecClientCert,
			cfg.Kafka.SecClientKey,
			cfg.Kafka.SecSkipVerify,
		)
	}
	kafkaProducer, err := kafka.NewProducer(ctx, pConfig)
	if err != nil {
		log.Fatal(ctx, "fatal error trying to create kafka producer", err, log.Data{"topic": cfg.Kafka.ContentUpdatedTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.LogErrors(ctx)

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
		if err := kafkaProducer.Initialise(ctx); err != nil {
			log.Warn(ctx, "failed to initialise kafka producer")
			return
		}
		kafkaProducer.Channels().Output <- bytes
	}
}

// scanEvent creates a ContentPublished event according to the user input
func scanEvent(scanner *bufio.Scanner) *models.ContentPublished {
	fmt.Println("--- [Send Kafka ContentPublished] ---")

	fmt.Println("Please type the URI")
	fmt.Printf("$ ")
	scanner.Scan()
	uri := scanner.Text()

	fmt.Println("Please type the dataset type (legacy or datasets)")
	fmt.Printf("$ ")
	scanner.Scan()
	dataType := scanner.Text()

	fmt.Println("Please type the collection ID")
	fmt.Printf("$ ")
	scanner.Scan()
	collectionID := scanner.Text()

	return &models.ContentPublished{
		URI:          uri,
		DataType:     dataType,
		TraceID:      "054435ded",
		CollectionID: collectionID,
	}
}
