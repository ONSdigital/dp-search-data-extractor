package private

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-search-data-extractor/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// UpdateTopicCache is a function to update the topic cache in publishing (private) mode.
// This function talks to the dp-topic-api via its private endpoints to retrieve the root topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.PrivateSubtopics which is then transformed in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty topic is returned
func UpdateTopicCache(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter) func() *cache.Topic {
	return func() *cache.Topic {
		processedTopics := make(map[string]struct{})

		// get root topic from dp-topic-api
		rootTopic, err := topicClient.GetRootTopicsPrivate(ctx, topicCli.Headers{ServiceAuthToken: "Bearer " + serviceAuthToken})
		if err != nil {
			logData := log.Data{
				"req_headers": topicCli.Headers{},
			}
			log.Error(ctx, "failed to get root topic from topic-api", err, logData)
			return cache.GetEmptyTopic()
		}

		// dereference root topics items to allow ranging through them
		var rootTopicItems []models.TopicResponse
		if rootTopic.PrivateItems != nil {
			rootTopicItems = *rootTopic.PrivateItems
		} else {
			err := errors.New("root topic private items is nil")
			log.Error(ctx, "failed to dereference root topic items pointer", err)
			return cache.GetEmptyTopic()
		}

		// Initialize topicCache
		topicCache := &cache.Topic{
			ID:              cache.TopicCacheKey,
			LocaliseKeyName: "Root",
			List:            cache.NewSubTopicsMap(),
		}

		// recursively process topics and their subtopics
		for i := range rootTopicItems {
			processTopic(ctx, serviceAuthToken, topicClient, rootTopicItems[i].ID, topicCache, processedTopics, "", 0)
		}

		// Check if any topics were found
		if len(topicCache.List.GetSubtopics()) == 0 {
			err := errors.New("root topic found, but no subtopics were returned")
			log.Error(ctx, "No topics loaded into cache - root topic found, but no subtopics were returned", err)
			return cache.GetEmptyTopic()
		}
		return topicCache
	}
}

func processTopic(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter, topicID string, topicCache *cache.Topic, processedTopics map[string]struct{}, parentTopicID string, depth int) {
	log.Info(ctx, "Processing topic at depth", log.Data{
		"topic_id": topicID,
		"depth":    depth,
	})

	// Check if the topic has already been processed
	if _, exists := processedTopics[topicID]; exists {
		err := errors.New("topic already processed")
		log.Error(ctx, "Skipping already processed topic", err, log.Data{
			"topic_id": topicID,
			"depth":    depth,
		})
		return
	}

	// Get the topic details from the topic client
	topic, err := topicClient.GetTopicPrivate(ctx, topicCli.Headers{ServiceAuthToken: "Bearer " + serviceAuthToken}, topicID)
	if err != nil {
		log.Error(ctx, "failed to get topic details from topic-api", err, log.Data{
			"topic_id": topicID,
			"depth":    depth,
		})
		return
	}

	if topic != nil {
		// Initialize subtopic list for the current topic if it doesn't exist
		subtopic := mapTopicModelToCache(*topic.Current, parentTopicID)

		// Add the current topic to the topicCache's List
		topicCache.List.AppendSubtopicID(topic.Current.Slug, subtopic)

		// Mark this topic as processed
		processedTopics[topicID] = struct{}{}

		// Process each subtopic recursively
		if topic.Current.SubtopicIds != nil {
			for _, subTopicID := range *topic.Current.SubtopicIds {
				processTopic(ctx, serviceAuthToken, topicClient, subTopicID, topicCache, processedTopics, topicID, depth+1)
			}
		}
	}
}

func mapTopicModelToCache(topic models.Topic, parentID string) cache.Subtopic {
	return cache.Subtopic{
		ID:              topic.ID,
		Slug:            topic.Slug,
		LocaliseKeyName: topic.Title,
		ReleaseDate:     topic.ReleaseDate,
		ParentID:        parentID,
	}
}
