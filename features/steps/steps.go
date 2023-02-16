package steps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	zClientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/dp-search-data-extractor/service"
	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
	"github.com/rdumont/assistdog"
)

func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I send a kafka event to content published topic$`, c.sendKafkafkaEvent)
	ctx.Step(`^The kafka event is processed$`, c.processKafkaEvent)
	ctx.Step(`^I should receive a kafka event to search-data-import topic with the following fields$`, c.iShouldReceiveKafkaEvent)
}

func (c *Component) sendKafkafkaEvent(table *godog.Table) error {
	observationEvents, err := c.convertToKafkaEvents(table)
	if err != nil {
		return err
	}

	for _, e := range observationEvents {
		if err := c.sendToConsumer(e); err != nil {
			return err
		}
	}

	return nil
}

func (c *Component) processKafkaEvent() error {
	c.inputData = models.ZebedeeData{
		DataType: "legacy",
		Description: models.Description{
			CDID:      "123",
			DatasetID: "456",
			Edition:   "something",
		},
	}

	marshelledData, err := json.Marshal(&c.inputData)
	if err != nil {
		return fmt.Errorf("error marshalling input data")
	}
	c.zebedeeClient = &zClientMock.ZebedeeClientMock{
		CheckerFunc: func(in1 context.Context, in2 *healthcheck.CheckState) error { return nil },
		GetPublishedDataFunc: func(ctx context.Context, uriString string) ([]byte, error) {
			return marshelledData, nil
		},
	}

	datasetAPIJSONResponse := dataset.Metadata{
		Version: dataset.Version{
			ReleaseDate: "releasedate",
		},
		DatasetDetails: dataset.DatasetDetails{
			Title:       "title",
			Description: "description",
			Keywords:    &[]string{"keyword1", "keyword2"},
		},
	}
	c.datasetClient = &zClientMock.DatasetClientMock{
		CheckerFunc: func(in1 context.Context, in2 *healthcheck.CheckState) error { return nil },
		GetVersionMetadataFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionId, datasetId, edition, version string) (dataset.Metadata, error) {
			return datasetAPIJSONResponse, nil
		},
	}
	// run application in separate goroutine
	go func() {
		c.svc, err = service.Run(context.Background(), c.serviceList, "", "", "", c.errorChan)
	}()

	signals := registerInterrupt()

	// kill application
	signals <- os.Interrupt

	return err
}

func (c *Component) iShouldReceiveKafkaEvent(events *godog.Table) error {
	assist := assistdog.NewDefault()
	assist.RegisterParser([]string{}, arrayParser)
	expected, err := assist.CreateSlice(new(models.SearchDataImport), events)
	if err != nil {
		return fmt.Errorf("failed to create slice from godog table: %w", err)
	}

	var got []*models.SearchDataImport

	select {
	case <-time.After(c.waitEventTimeout):
		return errors.New("timeout while waiting for kafka message being produced")
	case <-c.KafkaProducer.Channels().Closer:
		return errors.New("closer channel closed")
	case msg, ok := <-c.KafkaProducer.Channels().Output:
		if !ok {
			return errors.New("upstream channel closed")
		}

		var e = &models.SearchDataImport{}
		if err := schema.SearchDataImportEvent.Unmarshal(msg, e); err != nil {
			return fmt.Errorf("error unmarshalling message: %w", err)
		}
		got = append(got, e)
	}

	if diff := cmp.Diff(got, expected); diff != "" {
		//nolint
		return fmt.Errorf("-got +expected)\n%s\n", diff)
	}
	return c.ErrorFeature.StepError()
}

func (c *Component) convertToKafkaEvents(table *godog.Table) ([]*models.ContentPublished, error) {
	assist := assistdog.NewDefault()
	events, err := assist.CreateSlice(&models.ContentPublished{}, table)
	if err != nil {
		return nil, err
	}
	return events.([]*models.ContentPublished), nil
}

func (c *Component) sendToConsumer(e *models.ContentPublished) error {
	bytes, err := schema.ContentPublishedEvent.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg, err := kafkatest.NewMessage(bytes, 0)
	if err != nil {
		return fmt.Errorf("failed to create kafka message: %w", err)
	}

	c.KafkaConsumer.Channels().Upstream <- msg
	return nil
}

func registerInterrupt() chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	return signals
}

// we are passing the string array as [xxxx,yyyy,zzz]
// this is required to support array being used in kafka messages
func arrayParser(raw string) (interface{}, error) {
	// remove the starting and trailing brackets
	str := strings.Trim(raw, "[]")
	if str == "" {
		return []string{}, nil
	}

	strArray := strings.Split(str, ",")
	return strArray, nil
}
