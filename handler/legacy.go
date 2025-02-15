package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-search-data-extractor/cache"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

// handleZebedeeType handles a kafka message corresponding to Zebedee event type (legacy)
func (h *ContentPublished) handleZebedeeType(ctx context.Context, cpEvent *models.ContentPublished) error {
	// obtain correct uri to callback to Zebedee to retrieve content metadata
	uri, InvalidURIErr := retrieveCorrectURI(cpEvent.URI)
	if InvalidURIErr != nil {
		return InvalidURIErr
	}

	failedCriteria := shouldNotBeIndexed(cpEvent)
	if failedCriteria != "" {
		log.Info(ctx, "not processing content as does not meet indexing criteria", log.Data{
			"criteria":  failedCriteria,
			"data_type": cpEvent.DataType,
			"URI":       cpEvent.URI,
		})
		return nil
	}

	zebedeeContentPublished, err := h.ZebedeeCli.GetPublishedData(ctx, uri)
	if err != nil {
		log.Error(ctx, "failed to retrieve published data from zebedee", err)
		return err
	}

	// byte slice to Json & unMarshall Json
	var zebedeeData models.ZebedeeData
	err = json.Unmarshal(zebedeeContentPublished, &zebedeeData)
	if err != nil {
		log.Fatal(ctx, "error while attempting to unmarshal zebedee response into zebedeeData", err)
		return err
	}

	// keywords validation
	log.Info(ctx, "zebedee data ", log.Data{
		"uid":           zebedeeData.UID,
		"keywords":      zebedeeData.Description.Keywords,
		"keywordsLimit": h.Cfg.KeywordsLimit,
	})

	if zebedeeData.Description.Title == "" {
		log.Info(ctx, "not processing content as no title present", log.Data{
			"UID": zebedeeData.UID,
			"URI": zebedeeData.URI,
		})
		return nil
	}

	// Map data returned by Zebedee to the kafka Event structure
	searchData := models.MapZebedeeDataToSearchDataImport(zebedeeData, h.Cfg.KeywordsLimit)
	searchData.TraceID = cpEvent.TraceID
	searchData.JobID = cpEvent.JobID
	searchData.SearchIndex = getIndexName(cpEvent.SearchIndex)
	if h.Cfg.EnableTopicTagging {
		searchData = tagSearchDataWithURITopics(ctx, searchData, h.Cache.Topic)
	}

	if err := h.Producer.Send(schema.SearchDataImportEvent, &searchData); err != nil {
		log.Error(ctx, "error while attempting to send SearchDataImport event to producer", err)
		return fmt.Errorf("failed to send search data import event: %w", err)
	}

	return nil
}

func shouldNotBeIndexed(cpEvent *models.ContentPublished) string {
	if strings.Contains(cpEvent.URI, "/timeseries/") && strings.Contains(cpEvent.URI, "/previous/") {
		return "Item is a previous timeseries"
	}

	return ""
}

// Auto tag search data topics based on the URI segments of the request
func tagSearchDataWithURITopics(ctx context.Context, searchData models.SearchDataImport, topicCache *cache.TopicCache) models.SearchDataImport {
	// Set to track unique topic IDs
	uniqueTopics := make(map[string]struct{})

	// Add existing topics in searchData.Topics
	for _, topicID := range searchData.Topics {
		uniqueTopics[topicID] = struct{}{}
	}

	// Break URI into segments and exclude the last segment
	uriSegments := strings.Split(searchData.URI, "/")

	// Add topics based on URI segments
	parentSlug := ""
	for _, segment := range uriSegments {
		AddTopicWithParents(ctx, segment, parentSlug, topicCache, uniqueTopics)
		parentSlug = segment // Update parentSlug for the next iteration
	}

	// Convert set to slice
	searchData.Topics = make([]string, 0, len(uniqueTopics))
	for topicID := range uniqueTopics {
		searchData.Topics = append(searchData.Topics, topicID)
	}

	return searchData
}

// AddTopicWithParents adds a topic and its parents to the uniqueTopics map if they don't already exist.
// It recursively adds parent topics until it reaches the root topic.
func AddTopicWithParents(ctx context.Context, slug, parentSlug string, topicCache *cache.TopicCache, uniqueTopics map[string]struct{}) {
	if topic, _ := topicCache.GetTopic(ctx, slug); topic != nil {
		// Find the topic by matching both slug and parentSlug
		if topic.Slug == slug && topic.ParentSlug == parentSlug {
			if _, alreadyProcessed := uniqueTopics[topic.ID]; !alreadyProcessed {
				uniqueTopics[topic.ID] = struct{}{}
				if topic.ParentSlug != "" {
					// Recursively add the parent topic
					AddTopicWithParents(ctx, topic.ParentSlug, "", topicCache, uniqueTopics)
				}
			}
			return // Stop processing once the correct topic is found and added
		}
	}
}

func retrieveCorrectURI(uri string) (correctURI string, err error) {
	correctURI = uri

	// Remove edition segment of path from Zebedee dataset uri to
	// enable retrieval of the dataset resource for edition
	if strings.Contains(uri, DatasetDataType) {
		correctURI, err = extractDatasetURI(uri)
		if err != nil {
			return "", err
		}
	}

	return correctURI, nil
}

func extractDatasetURI(editionURI string) (string, error) {
	parsedURI, err := url.Parse(editionURI)
	if err != nil {
		return "", err
	}

	slicedURI := strings.Split(parsedURI.Path, "/")
	slicedURI = slicedURI[:len(slicedURI)-1]
	datasetURI := strings.Join(slicedURI, "/")

	return datasetURI, nil
}
