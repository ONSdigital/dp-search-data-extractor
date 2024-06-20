package private

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-search-data-extractor/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// UpdateDataTopics is a function to update the data topic cache in publishing (private) mode.
// This function talks to the dp-topic-api via its private endpoints to retrieve the root topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.PrivateSubtopics which is then transformed in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty topic is returned
func UpdateDataTopics(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter) func() []*cache.Topic {
	return func() []*cache.Topic {
		var topics []*cache.Topic
		processedTopics := make(map[string]bool)

		// get root topic from dp-topic-api
		rootTopic, err := topicClient.GetRootTopicsPrivate(ctx, topicCli.Headers{ServiceAuthToken: "Bearer " + serviceAuthToken})
		if err != nil {
			logData := log.Data{
				"req_headers": topicCli.Headers{},
			}
			log.Error(ctx, "failed to get root topic from topic-api", err, logData)
			return []*cache.Topic{cache.GetEmptyTopic()}
		}

		// deference rootTopic's subtopics(items) to allow ranging through them
		var rootTopicItems []models.TopicResponse
		if rootTopic.PrivateItems != nil {
			rootTopicItems = *rootTopic.PrivateItems
		} else {
			err := errors.New("root topic subtopics(items) is nil")
			log.Error(ctx, "failed to deference rootTopic subtopics pointer", err)
			return []*cache.Topic{cache.GetEmptyTopic()}
		}

		// recursively process topics and their subtopics
		for i := range rootTopicItems {
			processTopic(ctx, serviceAuthToken, topicClient, rootTopicItems[i].ID, &topics, processedTopics, "")
		}

		// Check if any data topics were found
		if len(topics) == 0 {
			err := errors.New("data root topic found, but no subtopics were returned")
			log.Error(ctx, "No topics loaded into cache - data root topic found, but no subtopics were returned", err)
			return []*cache.Topic{cache.GetEmptyTopic()}
		}
		return topics
	}
}

func processTopic(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter, topicID string, topics *[]*cache.Topic, processedTopics map[string]bool, parentTopicID string) {
	// Check if the topic is already processed
	if processedTopics[topicID] {
		return
	}

	// Get the topic details from the topic client
	dataTopic, err := topicClient.GetTopicPrivate(ctx, topicCli.Headers{ServiceAuthToken: serviceAuthToken}, topicID)
	if err != nil {
		log.Error(ctx, "failed to get topic details from topic-api", err)
		return
	}

	if dataTopic != nil {
		// Append the current topic to the list of topics
		// subtopicsChan := make(chan models.Topic)
		*topics = append(*topics, mapTopicModelToCache(*dataTopic.Current, parentTopicID))
		// Mark this topic as processed
		processedTopics[topicID] = true

		// Process each subtopic recursively
		if dataTopic.Current.SubtopicIds != nil {
			for _, subTopicID := range *dataTopic.Current.SubtopicIds {
				processTopic(ctx, serviceAuthToken, topicClient, subTopicID, topics, processedTopics, topicID)
			}
		}
	}
}

func mapTopicModelToCache(rootTopic models.Topic, parentID string) *cache.Topic {
	rootTopicCache := &cache.Topic{
		ID:              rootTopic.ID,
		Slug:            rootTopic.Slug,
		LocaliseKeyName: rootTopic.Title,
		ParentID:        parentID,
		ReleaseDate:     rootTopic.ReleaseDate,
		List:            cache.NewSubTopicsMap(),
	}
	return rootTopicCache
}
