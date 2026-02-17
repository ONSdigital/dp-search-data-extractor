package steps

import (
	"fmt"
	"net/http"

	"github.com/cucumber/godog"
)

const ContentUpdatedTopic = "content-updated"
const SearchContentUpdatedTopic = "search-content-updated"

func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the service starts`, c.theServiceStarts)
	ctx.Step(`^dp-dataset-api is healthy`, c.datasetAPIIsHealthy)
	ctx.Step(`^dp-dataset-api is unhealthy`, c.datasetAPIIsUnhealthy)
	ctx.Step(`^zebedee is healthy`, c.zebedeeIsHealthy)
	ctx.Step(`^zebedee is unhealthy`, c.zebedeeIsUnhealthy)
	ctx.Step(`^the following published data for uri "([^"]*)" is available in zebedee$`, c.theFollowingZebedeeResponseIsAvailable)
	ctx.Step(`^the following metadata with dataset-id "([^"]*)", edition "([^"]*)" and version "([^"]*)" is available in dp-dataset-api$`, c.theFollowingDatasetMetadataResponseIsAvailable)
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
