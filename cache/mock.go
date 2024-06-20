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

	testTopicCache.Set("economy", GetMockTopic("6734", "economy", "Economy", ""))
	testTopicCache.Set("environmentalaccounts", GetMockTopic("1834", "environmentalaccounts", "Environmental Accounts", "6734"))
	testTopicCache.Set("governmentpublicsectorandtaxes", GetMockTopic("8268", "governmentpublicsectorandtaxes", "Government Public Sector and Taxes", "6734"))
	testTopicCache.Set("publicsectorfinance", GetMockTopic("3687", "publicsectorfinance", "Public Sector Finance", "8268"))

	return testTopicCache, nil
}

// GetMockTopic returns a mocked topic which contains all the information for the mock data topic
func GetMockTopic(id, slug, title, parentID string) *Topic {
	mockTopic := &Topic{
		ID:              id,
		Slug:            slug,
		LocaliseKeyName: title,
		ParentID:        parentID,
		Query:           "1834,1234",
	}

	mockTopic.List = NewSubTopicsMap()
	mockTopic.List.AppendSubtopicID("1234", Subtopic{ID: "1234", LocaliseKeyName: "International Migration", ReleaseDate: timeHelper("2022-10-10T08:30:00Z")})
	mockTopic.List.AppendSubtopicID("1834", Subtopic{ID: "1834", LocaliseKeyName: "Environmental Accounts", ReleaseDate: timeHelper("2022-11-09T09:30:00Z")})

	return mockTopic
}

// timeHelper is a helper function given a time returns a Time pointer
func timeHelper(timeFormat string) *time.Time {
	t, _ := time.Parse(time.RFC3339, timeFormat)
	return &t
}
