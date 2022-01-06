package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	zClientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/dp-search-data-extractor/service"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
)

func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I send a kafka event to content published topic$`, c.sendKafkafkaEvent)
	ctx.Step(`^The kafka event is processed$`, c.processKafkaEvent)
	ctx.Step(`^I should receive the published data$`, c.iShouldReceivePublishedData)
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
		DataType: "Reviewed-uris",
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

func (c *Component) iShouldReceivePublishedData() error {
	outputData := <-c.KafkaProducer.Channels().Output

	var data = &models.SearchDataImport{}

	_ = schema.SearchDataImportEvent.Unmarshal(outputData, data)

	assert.Equal(&c.ErrorFeature, c.inputData.Description.CDID, data.CDID)
	assert.Equal(&c.ErrorFeature, c.inputData.Description.DatasetID, data.DatasetID)
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
		return err
	}

	c.KafkaConsumer.Channels().Upstream <- kafkatest.NewMessage(bytes, 0)
	return nil
}

func registerInterrupt() chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	return signals
}
