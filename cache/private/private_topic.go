package private

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-search-data-extractor/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// UpdateTopics is a function to update the topic cache in publishing (private) mode.
// This function talks to the dp-topic-api via its private endpoints to retrieve the root topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.PrivateSubtopics which is then transformed in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty topic is returned
func UpdateTopics(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter) func() []*cache.Topic {
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

		// dereference rootTopic's subtopics(items) to allow ranging through them
		var rootTopicItems []models.TopicResponse
		if rootTopic.PrivateItems != nil {
			rootTopicItems = *rootTopic.PrivateItems
		} else {
			err := errors.New("root topic subtopics(items) is nil")
			log.Error(ctx, "failed to dereference rootTopic subtopics pointer", err)
			return []*cache.Topic{cache.GetEmptyTopic()}
		}

		// recursively process topics and their subtopics
		for i := range rootTopicItems {
			processTopic(ctx, serviceAuthToken, topicClient, rootTopicItems[i].ID, &topics, processedTopics, "")
		}

		// Check if any topics were found
		if len(topics) == 0 {
			err := errors.New("root topic found, but no subtopics were returned")
			log.Error(ctx, "No topics loaded into cache - root topic found, but no subtopics were returned", err)
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
	topic, err := topicClient.GetTopicPrivate(ctx, topicCli.Headers{ServiceAuthToken: "Bearer " + serviceAuthToken}, topicID)
	if err != nil {
		logData := log.Data{
			topicID: topicID,
		}
		log.Error(ctx, "failed to get topic details from topic-api", err, logData)
		return
	}

	if topic != nil {
		// Append the current topic to the list of topics
		*topics = append(*topics, mapTopicModelToCache(*topic.Current, parentTopicID))
		// Mark this topic as processed
		processedTopics[topicID] = true

		// Process each subtopic recursively
		if topic.Current.SubtopicIds != nil {
			for _, subTopicID := range *topic.Current.SubtopicIds {
				processTopic(ctx, serviceAuthToken, topicClient, subTopicID, topics, processedTopics, topicID)
			}
		}
	}
}

func mapTopicModelToCache(topic models.Topic, parentID string) *cache.Topic {
	rootTopicCache := &cache.Topic{
		ID:              topic.ID,
		Slug:            topic.Slug,
		LocaliseKeyName: topic.Title,
		ParentID:        parentID,
		ReleaseDate:     topic.ReleaseDate,
		List:            cache.NewSubTopicsMap(),
	}
	return rootTopicCache
}
