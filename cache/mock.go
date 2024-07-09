package cache

import (
	"context"
	"time"
)

// GetMockCacheList returns a mocked list of cache which contains the topic cache
func GetMockCacheList(ctx context.Context) (*List, error) {
	testTopicCache, err := getMockTopicCache(ctx)
	if err != nil {
		return nil, err
	}

	cacheList := List{
		Topic: testTopicCache,
	}

	return &cacheList, nil
}

// getMockTopicCache returns a mocked topic cache which contains all the information for the mock topics
func getMockTopicCache(ctx context.Context) (*TopicCache, error) {
	testTopicCache, err := NewTopicCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	rootTopicID := testTopicCache.GetTopicCacheKey()
	testTopicCache.Set(rootTopicID, GetMockRootTopic(rootTopicID))

	return testTopicCache, nil
}

// GetMockRootTopic returns the mocked root topic
func GetMockRootTopic(rootTopicID string) *Topic {
	mockRootTopic := &Topic{
		ID:   rootTopicID,
		Slug: "root",
	}

	mockRootTopic.List = NewSubTopicsMap()
	mockRootTopic.List.AppendSubtopicID("economy", Subtopic{ID: "6734", Slug: "economy", LocaliseKeyName: "Economy", ReleaseDate: timeHelper("2022-10-10T08:30:00Z"), ParentID: ""})
	mockRootTopic.List.AppendSubtopicID("environmentalaccounts", Subtopic{ID: "1834", Slug: "environmentalaccounts", LocaliseKeyName: "Environmental Accounts", ReleaseDate: timeHelper("2022-10-10T08:30:00Z"), ParentID: "6734"})
	mockRootTopic.List.AppendSubtopicID("governmentpublicsectorandtaxes", Subtopic{ID: "8268", Slug: "governmentpublicsectorandtaxes", LocaliseKeyName: "Government Public Sector and Taxes", ReleaseDate: timeHelper("2022-10-10T08:30:00Z"), ParentID: "6734"})
	mockRootTopic.List.AppendSubtopicID("publicsectorfinance", Subtopic{ID: "3687", Slug: "publicsectorfinance", LocaliseKeyName: "Public Sector Finance", ReleaseDate: timeHelper("2022-10-10T08:30:00Z"), ParentID: "8268"})
	mockRootTopic.List.AppendSubtopicID("internationalmigration", Subtopic{ID: "1234", Slug: "internationalmigration", LocaliseKeyName: "International Migration", ReleaseDate: timeHelper("2022-10-10T08:30:00Z")})

	return mockRootTopic
}

// timeHelper is a helper function given a time returns a Time pointer
func timeHelper(timeFormat string) *time.Time {
	t, _ := time.Parse(time.RFC3339, timeFormat)
	return &t
}
