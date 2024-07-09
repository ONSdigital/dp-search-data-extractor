package private

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-search-data-extractor/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/dp-topic-api/sdk"
	topicCliErr "github.com/ONSdigital/dp-topic-api/sdk/errors"
	mockTopic "github.com/ONSdigital/dp-topic-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testEconomyRootTopic = models.Topic{
		ID:          "6734",
		Title:       "Economy",
		SubtopicIds: &[]string{"1834"},
	}

	testEconomyRootTopic2 = models.Topic{
		ID:          "6734",
		Title:       "Economy",
		SubtopicIds: &[]string{"1834", "1234"},
	}

	testBusinessRootTopic = models.Topic{
		ID:          "1234",
		Title:       "Business",
		SubtopicIds: &[]string{},
	}

	testBusinessRootTopic2 = models.Topic{
		ID:          "1234",
		Title:       "Business",
		SubtopicIds: &[]string{"6734"},
	}

	testEconomyRootTopicPrivate = models.TopicResponse{
		ID:      "6734",
		Next:    &testEconomyRootTopic,
		Current: &testEconomyRootTopic,
	}

	testEconomyRootTopicPrivate2 = models.TopicResponse{
		ID:      "6734",
		Next:    &testEconomyRootTopic2,
		Current: &testEconomyRootTopic2,
	}

	testBusinessRootTopicPrivate = models.TopicResponse{
		ID:      "1234",
		Next:    &testBusinessRootTopic,
		Current: &testBusinessRootTopic,
	}

	testBusinessRootTopicPrivate2 = models.TopicResponse{
		ID:      "1234",
		Next:    &testBusinessRootTopic2,
		Current: &testBusinessRootTopic2,
	}

	expectedTopicCache = &cache.Topic{
		ID:              cache.TopicCacheKey,
		LocaliseKeyName: "Root",
		Query:           "6734, 1234",
	}
)

func TestUpdateTopicCache(t *testing.T) {
	ctx := context.Background()
	serviceAuthToken := "test-token"

	emptyTopic := cache.GetEmptyTopic()

	Convey("Given root topic exist and have subtopics with no duplicate topics", t, func() {
		mockClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				return &models.PrivateSubtopics{
					Count:        2,
					Offset:       0,
					Limit:        50,
					TotalCount:   2,
					PrivateItems: &[]models.TopicResponse{testEconomyRootTopicPrivate, testBusinessRootTopicPrivate},
				}, nil
			},
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.TopicResponse, topicCliErr.Error) {
				switch topicID {
				case "6734":
					return &models.TopicResponse{
						ID: topicID,
						Current: &models.Topic{
							ID:          "6734",
							Title:       "Economy",
							Slug:        "economy",
							SubtopicIds: &[]string{"1834"},
						},
					}, nil
				case "1834":
					return &models.TopicResponse{
						ID: topicID,
						Current: &models.Topic{
							ID:          "1834",
							Title:       "Environmental Accounts",
							Slug:        "environmentalaccounts",
							SubtopicIds: &[]string{},
						},
					}, nil
				case "1234":
					return &models.TopicResponse{
						ID: topicID,
						Current: &models.Topic{
							ID:          "1234",
							Title:       "Business",
							Slug:        "business",
							SubtopicIds: &[]string{},
						},
					}, nil
				default:
					return nil, topicCliErr.StatusError{
						Err: errors.New("unexpected error"),
					}
				}
			},
		}

		Convey("When UpdateTopicCache (Private) is called", func() {
			respTopic := UpdateTopicCache(ctx, serviceAuthToken, mockClient)()

			Convey("Then the topics cache is returned", func() {
				So(respTopic, ShouldNotBeNil)
				So(respTopic.ID, ShouldEqual, expectedTopicCache.ID)
				So(respTopic.LocaliseKeyName, ShouldEqual, expectedTopicCache.LocaliseKeyName)

				So(len(respTopic.List.GetSubtopics()), ShouldEqual, 3)
			})
		})

		Convey("And given root topics exist and have subtopics with duplicate topics", func() {
			mockClient.GetRootTopicsPrivateFunc = func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				return &models.PrivateSubtopics{
					Count:        2,
					Offset:       0,
					Limit:        50,
					TotalCount:   2,
					PrivateItems: &[]models.TopicResponse{testEconomyRootTopicPrivate, testBusinessRootTopicPrivate, testEconomyRootTopicPrivate, testBusinessRootTopicPrivate},
				}, nil
			}

			Convey("When UpdateTopicCache (Private) is called", func() {
				respTopic := UpdateTopicCache(ctx, serviceAuthToken, mockClient)()

				Convey("Then the topics cache is returned with expected number of topics excluding duplicates", func() {
					So(respTopic, ShouldNotBeNil)
					So(respTopic.ID, ShouldEqual, expectedTopicCache.ID)
					So(respTopic.LocaliseKeyName, ShouldEqual, expectedTopicCache.LocaliseKeyName)

					So(len(respTopic.List.GetSubtopics()), ShouldEqual, 3)
				})
			})
		})

		Convey("And given root topics exist and private items that have subtopics that can end up in a recursive loop", func() {
			mockClient.GetRootTopicsPrivateFunc = func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				return &models.PrivateSubtopics{
					Count:        2,
					Offset:       0,
					Limit:        50,
					TotalCount:   2,
					PrivateItems: &[]models.TopicResponse{testEconomyRootTopicPrivate, testEconomyRootTopicPrivate2, testBusinessRootTopicPrivate2},
				}, nil
			}

			Convey("When UpdateTopicCache (Private) is called", func() {
				respTopic := UpdateTopicCache(ctx, serviceAuthToken, mockClient)()

				Convey("Then the topics cache is returned as expected with no duplicates and does not get stuck in a loop", func() {
					So(respTopic, ShouldNotBeNil)
					So(respTopic.ID, ShouldEqual, expectedTopicCache.ID)
					So(respTopic.LocaliseKeyName, ShouldEqual, expectedTopicCache.LocaliseKeyName)

					So(len(respTopic.List.GetSubtopics()), ShouldEqual, 3)
				})
			})
		})
	})

	Convey("Given an error in getting root topics from topic-api", t, func() {
		mockClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When UpdateTopicCache (Private) is called", func() {
			respTopic := UpdateTopicCache(ctx, serviceAuthToken, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopic, ShouldResemble, emptyTopic)
			})
		})
	})

	Convey("Given root topics private items is nil", t, func() {
		mockClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				return &models.PrivateSubtopics{
					Count:        1,
					Offset:       0,
					Limit:        50,
					TotalCount:   1,
					PrivateItems: nil,
				}, nil
			},
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.TopicResponse, topicCliErr.Error) {
				return &models.TopicResponse{
					Current: &models.Topic{
						ID:          topicID,
						SubtopicIds: nil,
					},
				}, nil
			},
		}

		Convey("When UpdateTopicCache (Private) is called", func() {
			respTopic := UpdateTopicCache(ctx, serviceAuthToken, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopic, ShouldResemble, emptyTopic)
			})
		})
	})

	Convey("Given root topics exist but no topics found", t, func() {
		mockClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				return &models.PrivateSubtopics{
					Count:        0,
					Offset:       0,
					Limit:        50,
					TotalCount:   0,
					PrivateItems: &[]models.TopicResponse{},
				}, nil
			},
		}

		Convey("When UpdateTopicCache (Private) is called", func() {
			respTopics := UpdateTopicCache(ctx, serviceAuthToken, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopics, ShouldResemble, emptyTopic)
			})
		})
	})
}
