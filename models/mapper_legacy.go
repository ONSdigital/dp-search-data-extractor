package models

import (
	"context"
	"strings"

	"github.com/ONSdigital/dp-search-data-extractor/cache"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	ReleaseDataType = "release"
)

// MapZebedeeDataToSearchDataImport Performs default mapping of zebedee data to a SearchDataImport struct.
// It also optionally takes a limit which truncates the keywords to the desired amount. This value can be -1 for no
// truncation.
func MapZebedeeDataToSearchDataImport(ctx context.Context, zebedeeData ZebedeeData, keywordsLimit int, topicCache *cache.TopicCache, enableTopicTagging bool) SearchDataImport {
	searchData := SearchDataImport{
		UID:             zebedeeData.URI,
		URI:             zebedeeData.URI,
		DataType:        zebedeeData.DataType,
		CDID:            zebedeeData.Description.CDID,
		DatasetID:       zebedeeData.Description.DatasetID,
		Keywords:        RectifyKeywords(zebedeeData.Description.Keywords, keywordsLimit),
		MetaDescription: zebedeeData.Description.MetaDescription,
		Summary:         zebedeeData.Description.Summary,
		ReleaseDate:     zebedeeData.Description.ReleaseDate,
		Title:           zebedeeData.Description.Title,
		Topics:          zebedeeData.Description.Topics,
	}
	if zebedeeData.Description.Edition != "" {
		searchData.Edition = zebedeeData.Description.Edition
	}
	if zebedeeData.DataType == ReleaseDataType {
		logData := log.Data{
			"zebedeeRCData": zebedeeData,
		}
		log.Info(context.Background(), "zebedee release calender data", logData)
		for _, data := range zebedeeData.DateChanges {
			searchData.DateChanges = append(searchData.DateChanges, ReleaseDateDetails(data))
		}
		searchData.Published = zebedeeData.Description.Published
		searchData.Cancelled = zebedeeData.Description.Cancelled
		searchData.Finalised = zebedeeData.Description.Finalised
		searchData.ProvisionalDate = zebedeeData.Description.ProvisionalDate
		searchData.Survey = zebedeeData.Description.Survey
		searchData.Language = zebedeeData.Description.Language
		searchData.CanonicalTopic = zebedeeData.Description.CanonicalTopic
	}

	if enableTopicTagging {
		// Set to track unique topic IDs
		uniqueTopics := make(map[string]struct{})

		// Add existing topics in searchData.Topics
		for _, topicID := range searchData.Topics {
			if _, exists := uniqueTopics[topicID]; !exists {
				uniqueTopics[topicID] = struct{}{}
			}
		}

		// Break URI into segments and exclude the last segment
		uriSegments := strings.Split(zebedeeData.URI, "/")

		// Add topics based on URI segments
		for _, segment := range uriSegments {
			AddTopicWithParents(ctx, segment, topicCache, uniqueTopics)
		}

		// Convert set to slice
		searchData.Topics = make([]string, 0, len(uniqueTopics))
		for topicID := range uniqueTopics {
			searchData.Topics = append(searchData.Topics, topicID)
		}
	}
	return searchData
}

func AddTopicWithParents(ctx context.Context, slug string, topicCache *cache.TopicCache, uniqueTopics map[string]struct{}) {
	if topic, _ := topicCache.GetTopic(ctx, slug); topic != nil {
		if _, exists := uniqueTopics[topic.ID]; !exists {
			uniqueTopics[topic.ID] = struct{}{}
			if topic.ParentSlug != "" {
				AddTopicWithParents(ctx, topic.ParentSlug, topicCache, uniqueTopics)
			}
		}
	}
}

// RectifyKeywords sanitises a slice of keywords, splitting any that contain commas into seperate keywords and trimming
// any whitespace. It also optionally takes a limit which truncates the keywords to the desired amount. This value can
// be -1 for no truncation.
func RectifyKeywords(keywords []string, keywordsLimit int) []string {
	var strArray []string
	rectifiedKeywords := make([]string, 0)

	if keywordsLimit == 0 {
		return []string{""}
	}

	for i := range keywords {
		strArray = strings.Split(keywords[i], ",")

		for j := range strArray {
			keyword := strings.TrimSpace(strArray[j])
			rectifiedKeywords = append(rectifiedKeywords, keyword)
		}
	}

	if (len(rectifiedKeywords) < keywordsLimit) || (keywordsLimit == -1) {
		return rectifiedKeywords
	}

	return rectifiedKeywords[:keywordsLimit]
}

func ValidateTopics(topics []string) []string {
	if topics == nil {
		topics = []string{""}
	}
	return topics
}
