package steps

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/dp-search-data-extractor/service"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
)

func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^these kafka events are consumed:$`, c.theseKafkaEventsAreConsumed)
	ctx.Step(`^I should receive a kafka response$`, c.iShouldReceiveAKafkaResponse)
}

func (c *Component) iShouldReceiveAKafkaResponse() error {
	content, err := ioutil.ReadFile(c.cfg.OutputFilePath)
	if err != nil {
		return err
	}

	assert.Equal(c, "Hello, testURL.com", string(content))

	return c.StepError()
}

func (c *Component) theseKafkaEventsAreConsumed(table *godog.Table) error {

	observationEvents, err := c.convertToKafkaEvents(table)
	if err != nil {
		return err
	}

	signals := registerInterrupt()

	// run application in separate goroutine
	go func() {
		c.svc, err = service.Run(context.Background(), c.serviceList, "", "", "", c.errorChan)
	}()

	// consume extracted observations
	for _, e := range observationEvents {
		if err := c.sendToConsumer(e); err != nil {
			return err
		}
	}

	time.Sleep(300 * time.Millisecond)

	// kill application
	signals <- os.Interrupt

	return nil
}

func (c *Component) convertToKafkaEvents(table *godog.Table) ([]*event.ContentPublished, error) {
	assist := assistdog.NewDefault()
	events, err := assist.CreateSlice(&event.ContentPublished{}, table)
	if err != nil {
		return nil, err
	}
	return events.([]*event.ContentPublished), nil
}

func (c *Component) sendToConsumer(e *event.ContentPublished) error {
	bytes, err := schema.ContentPublishedSchema.Marshal(e)
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
