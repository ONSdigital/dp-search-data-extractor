package steps

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
	"github.com/rdumont/assistdog"
)

func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	// ctx.Step(`^I send a kafka event to content published topic$`, c.sendKafkafkaEvent)
	ctx.Step(`^the service starts`, c.theServiceStarts)
	ctx.Step(`^dp-dataset-api is healthy`, c.datasetAPIIsHealthy)
	ctx.Step(`^dp-dataset-api is unhealthy`, c.datasetAPIIsUnhealthy)
	ctx.Step(`^zebedee is healthy`, c.zebedeeIsHealthy)
	ctx.Step(`^zebedee is unhealthy`, c.zebedeeIsUnhealthy)
	ctx.Step(`^the following published data for uri "([^"]*)" is available in zebedee$`, c.theFollowingZebedeeResponseIsAvailable)
	ctx.Step(`^the following metadata with dataset-id "([^"]*)", edition "([^"]*)" and version "([^"]*)" is available in dp-dataset-api$`, c.theFollowingDatasetMetadataResponseIsAvailable)
	ctx.Step(`^this content-updated event is queued, to be consumed$`, c.thisContentUpdatedEventIsQueued)
	ctx.Step(`^no search-data-import events are produced`, c.noEventsAreProduced)
	ctx.Step(`^this search-data-import event is sent$`, c.thisSearchDataImportEventIsSent)
}

// theServiceStarts starts the service under test in a new go-routine
// note that this step should be called only after all dependencies have been setup,
// to prevent any race condition, specially during the first healthcheck iteration.
func (c *Component) theServiceStarts() error {
	return c.startService(c.ctx)
}

// datasetAPIIsHealthy generates a mocked healthy response for dataset API healthecheck
func (c *Component) datasetAPIIsHealthy() error {
	const res = `{"status": "OK"}`
	c.DatasetAPI.NewHandler().
		Get("/health").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
}

// datasetAPIIsUnhealthy generates a mocked unhealthy response for dataset API healthcheck
func (c *Component) datasetAPIIsUnhealthy() error {
	const res = `{"status": "CRITICAL"}`
	c.DatasetAPI.NewHandler().
		Get("/health").
		Reply(http.StatusTooManyRequests).
		BodyString(res)
	return nil
}

// zebedeeIsHealthy generates a mocked healthy response for zebedee healthecheck
func (c *Component) zebedeeIsHealthy() error {
	const res = `{"status": "OK"}`
	c.Zebedee.NewHandler().
		Get("/health").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
}

// zebedeeIsUnhealthy generates a mocked unhealthy response for zebedee healthcheck
func (c *Component) zebedeeIsUnhealthy() error {
	const res = `{"status": "CRITICAL"}`
	c.Zebedee.NewHandler().
		Get("/health").
		Reply(http.StatusTooManyRequests).
		BodyString(res)
	return nil
}

// theFollowingZebedeeResponseIsAvailable generate a mocked response for zebedee
// GET /publisheddata?uri={uriString} with the provided response
func (c *Component) theFollowingZebedeeResponseIsAvailable(uriString string, zebedeeData *godog.DocString) error {
	c.Zebedee.NewHandler().
		Get("/publisheddata?uri=" + uriString).
		Reply(http.StatusOK).
		BodyString(zebedeeData.Content)

	return nil
}

// theFollowingDatasetMetadataResponseIsAvailable generate a mocked response for dataset API
// GET /dataset/{id}/editions/{edition}/versions/{version}/metadata with the provided metadata response
func (c *Component) theFollowingDatasetMetadataResponseIsAvailable(id, edition, version string, metadata *godog.DocString) error {
	c.DatasetAPI.NewHandler().
		Get(fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/metadata", id, edition, version)).
		Reply(http.StatusOK).
		BodyString(metadata.Content).
		AddHeader("Etag", c.testETag)

	return nil
}

// func (c *Component) sendKafkafkaEvent(table *godog.Table) error {
// 	observationEvents, err := c.convertToKafkaEvents(table)
// 	if err != nil {
// 		return err
// 	}

// 	for _, e := range observationEvents {
// 		if err := c.sendToConsumer(e); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// thisContentUpdatedEventIsQueued produces a new ContentPublished event with the contents defined by the input
func (c *Component) thisContentUpdatedEventIsQueued(table *godog.Table) error {
	assist := assistdog.NewDefault()
	assist.RegisterParser([]string{}, arrayParser)
	r, err := assist.CreateSlice(&models.ContentPublished{}, table)
	if err != nil {
		return fmt.Errorf("failed to create slice from godog table: %w", err)
	}
	expected := r.([]*models.ContentPublished)

	log.Info(c.ctx, "event to queue for testing: ", log.Data{
		"event": expected[0],
	})
	if err := c.KafkaConsumer.QueueMessage(schema.ContentPublishedEvent, expected[0]); err != nil {
		return fmt.Errorf("failed to queue event for testing: %w", err)
	}

	log.Info(c.ctx, "queue func is going to exit")
	return nil
}

// thisSearchDataImportEventIsSent checks that the provided event is produced and sent to kafka
func (c *Component) thisSearchDataImportEventIsSent(table *godog.Table) error {
	assist := assistdog.NewDefault()
	assist.RegisterParser([]string{}, arrayParser)
	r, err := assist.CreateSlice(new(models.SearchDataImport), table)
	if err != nil {
		return fmt.Errorf("failed to create slice from godog table: %w", err)
	}
	expected := r.([]*models.SearchDataImport)

	log.Info(c.ctx, "[DEBUG] expected set")

	var got []*models.SearchDataImport

	var e = &models.SearchDataImport{}
	if err := c.KafkaProducer.WaitForMessageSent(schema.SearchDataImportEvent, e, c.waitEventTimeout); err != nil {
		return fmt.Errorf("failed to expect sent message: %w", err)
	}
	got = append(got, e)

	log.Info(c.ctx, "[DEBUG] got", log.Data{"got": got})

	// select {
	// case <-time.After(c.waitEventTimeout):
	// 	return errors.New("timeout while waiting for kafka message being produced")
	// case <-c.svc.Producer.Channels().Closer:
	// 	return errors.New("closer channel closed")
	// case msg, ok := <-c.svc.Producer.Channels().Output:
	// 	if !ok {
	// 		return errors.New("upstream channel closed")
	// 	}

	// 	var e = &models.SearchDataImport{}
	// 	if err := schema.SearchDataImportEvent.Unmarshal(msg, e); err != nil {
	// 		return fmt.Errorf("error unmarshalling message: %w", err)
	// 	}
	// 	got = append(got, e)
	// }

	if diff := cmp.Diff(got, expected); diff != "" {
		//nolint
		return fmt.Errorf("-got +expected)\n%s\n", diff)
	}

	return c.ErrorFeature.StepError()
}

// noEventsAreProduced waits on the service's kafka producer output channel
// and validates that nothing is sent, within a time window of duration waitEventTimeout
func (c *Component) noEventsAreProduced() error {
	select {
	case <-time.After(c.waitEventTimeout):
		return nil
	case <-c.svc.Producer.Channels().Closer:
		return errors.New("closer channel closed")
	case msg, ok := <-c.svc.Producer.Channels().Output:
		if !ok {
			return errors.New("output channel closed")
		}

		var e = &models.SearchDataImport{}
		if err := schema.SearchDataImportEvent.Unmarshal(msg, e); err != nil {
			return fmt.Errorf("error unmarshalling message: %w", err)
		}

		return fmt.Errorf("kafka event received in csv-created topic: %v", e)
	}
}

// func (c *Component) processKafkaEvent() error {
// 	c.inputData = models.ZebedeeData{
// 		DataType: "legacy",
// 		Description: models.Description{
// 			CDID:      "123",
// 			DatasetID: "456",
// 			Edition:   "something",
// 		},
// 	}

// 	marshelledData, err := json.Marshal(&c.inputData)
// 	if err != nil {
// 		return fmt.Errorf("error marshalling input data")
// 	}
// 	c.zebedeeClient = &zClientMock.ZebedeeClientMock{
// 		CheckerFunc: func(in1 context.Context, in2 *healthcheck.CheckState) error { return nil },
// 		GetPublishedDataFunc: func(ctx context.Context, uriString string) ([]byte, error) {
// 			return marshelledData, nil
// 		},
// 	}

// 	datasetAPIJSONResponse := dataset.Metadata{
// 		Version: dataset.Version{
// 			ReleaseDate: "releasedate",
// 		},
// 		DatasetDetails: dataset.DatasetDetails{
// 			Title:       "title",
// 			Description: "description",
// 			Keywords:    &[]string{"keyword1", "keyword2"},
// 		},
// 	}
// 	c.datasetClient = &zClientMock.DatasetClientMock{
// 		CheckerFunc: func(in1 context.Context, in2 *healthcheck.CheckState) error { return nil },
// 		GetVersionMetadataFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionId, datasetId, edition, version string) (dataset.Metadata, error) {
// 			return datasetAPIJSONResponse, nil
// 		},
// 	}
// 	// run application in separate goroutine
// 	go func() {
// 		c.svc, err = service.Run(context.Background(), c.serviceList, "", "", "", c.errorChan)
// 	}()

// 	signals := registerInterrupt()

// 	// kill application
// 	signals <- os.Interrupt

// 	return err
// }

// func (c *Component) iShouldReceiveKafkaEvent(events *godog.Table) error {
// 	assist := assistdog.NewDefault()
// 	assist.RegisterParser([]string{}, arrayParser)
// 	expected, err := assist.CreateSlice(new(models.SearchDataImport), events)
// 	if err != nil {
// 		return fmt.Errorf("failed to create slice from godog table: %w", err)
// 	}

// 	var got []*models.SearchDataImport

// 	select {
// 	case <-time.After(c.waitEventTimeout):
// 		return errors.New("timeout while waiting for kafka message being produced")
// 	case <-c.svc.Producer.Channels().Closer:
// 		return errors.New("closer channel closed")
// 	case msg, ok := <-c.svc.Producer.Channels().Output:
// 		if !ok {
// 			return errors.New("upstream channel closed")
// 		}

// 		var e = &models.SearchDataImport{}
// 		if err := schema.SearchDataImportEvent.Unmarshal(msg, e); err != nil {
// 			return fmt.Errorf("error unmarshalling message: %w", err)
// 		}
// 		got = append(got, e)
// 	}

// 	if diff := cmp.Diff(got, expected); diff != "" {
// 		//nolint
// 		return fmt.Errorf("-got +expected)\n%s\n", diff)
// 	}
// 	return c.ErrorFeature.StepError()
// }

// noEventsAreProduced waits on the service's kafka producer output channel
// and validates that nothing is sent, within a time window of duration waitEventTimeout
// func (c *Component) noEventsAreProduced() error {
// 	select {
// 	case <-time.After(c.waitEventTimeout):
// 		return nil
// 	case <-c.svc.Producer.Channels().Closer:
// 		return errors.New("closer channel closed")
// 	case msg, ok := <-c.svc.Producer.Channels().Output:
// 		if !ok {
// 			return errors.New("output channel closed")
// 		}

// 		var e = &models.SearchDataImport{}
// 		if err := schema.SearchDataImportEvent.Unmarshal(msg, e); err != nil {
// 			return fmt.Errorf("error unmarshalling message: %w", err)
// 		}

// 		return fmt.Errorf("kafka event received in csv-created topic: %v", e)
// 	}
// }

// func (c *Component) convertToKafkaEvents(table *godog.Table) ([]*models.ContentPublished, error) {
// 	assist := assistdog.NewDefault()
// 	assist.RegisterParser([]string{}, arrayParser)
// 	events, err := assist.CreateSlice(&models.ContentPublished{}, table)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return events.([]*models.ContentPublished), nil
// }

// func (c *Component) sendToConsumer(e *models.ContentPublished) error {
// 	bytes, err := schema.ContentPublishedEvent.Marshal(e)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal event: %w", err)
// 	}

// 	msg, err := kafkatest.NewMessage(bytes, 0)
// 	if err != nil {
// 		return fmt.Errorf("failed to create kafka message: %w", err)
// 	}

// 	c.svc.Consumer.Channels().Upstream <- msg
// 	return nil
// }

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
