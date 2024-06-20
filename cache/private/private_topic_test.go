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
	testEconomyRootSubTopicPrivate = models.TopicResponse{
		ID:      "6734",
		Next:    &testEconomyRootTopic,
		Current: &testEconomyRootTopic,
	}

	testBusinessRootSubTopicPrivate = models.TopicResponse{
		ID:      "1234",
		Next:    &testBusinessRootTopic,
		Current: &testBusinessRootTopic,
	}

	testEconomyRootTopic = models.Topic{
		ID:          "6734",
		Title:       "Economy",
		SubtopicIds: &[]string{"1834"},
	}

	testBusinessRootTopic = models.Topic{
		ID:          "1234",
		Title:       "Business",
		SubtopicIds: &[]string{},
	}
)

func TestUpdateDataTopics(t *testing.T) {
	ctx := context.Background()
	serviceAuthToken := "test-token"

	expectedTopics := []*cache.Topic{
		{ID: "6734", LocaliseKeyName: "Economy", Query: "6734,1834"},
		{ID: "1834", LocaliseKeyName: "Environmental Accounts"},
		{ID: "1234", LocaliseKeyName: "Business"},
	}
	emptyTopic := []*cache.Topic{cache.GetEmptyTopic()}

	Convey("Given root topics exist and have subtopics", t, func() {
		mockClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				return &models.PrivateSubtopics{
					Count:        2,
					Offset:       0,
					Limit:        50,
					TotalCount:   2,
					PrivateItems: &[]models.TopicResponse{testEconomyRootSubTopicPrivate, testBusinessRootSubTopicPrivate},
				}, nil
			},
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.TopicResponse, topicCliErr.Error) {
				switch topicID {
				case "6734":
					return &models.TopicResponse{
						Current: &models.Topic{
							ID:          "6734",
							Title:       "Economy",
							SubtopicIds: &[]string{"1834"},
						},
					}, nil
				case "1834":
					return &models.TopicResponse{
						Current: &models.Topic{
							ID:          "1834",
							Title:       "Environmental Accounts",
							SubtopicIds: &[]string{},
						},
					}, nil
				case "1234":
					return &models.TopicResponse{
						Current: &models.Topic{
							ID:          "1234",
							Title:       "Business",
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

		Convey("When UpdateDataTopicsPrivate is called", func() {
			respTopics := UpdateDataTopics(ctx, serviceAuthToken, mockClient)()

			Convey("Then the topics cache is returned", func() {
				So(respTopics, ShouldNotBeNil)
				So(len(respTopics), ShouldEqual, len(expectedTopics))
				for i, expected := range expectedTopics {
					So(respTopics[i].ID, ShouldEqual, expected.ID)
					So(respTopics[i].LocaliseKeyName, ShouldEqual, expected.LocaliseKeyName)
				}
			})
		})
	})

	Convey("Given an error in getting root topics from topic-api", t, func() {
		mockClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				return &models.PrivateSubtopics{
					Count:        1,
					Offset:       0,
					Limit:        50,
					TotalCount:   1,
					PrivateItems: &[]models.TopicResponse{testEconomyRootSubTopicPrivate, testBusinessRootSubTopicPrivate},
				}, nil
			},
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.TopicResponse, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When UpdateDataTopicsPrivate is called", func() {
			respTopics := UpdateDataTopics(ctx, serviceAuthToken, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopics, ShouldResemble, emptyTopic)
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

		Convey("When UpdateDataTopicsPrivate is called", func() {
			respTopics := UpdateDataTopics(ctx, serviceAuthToken, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopics, ShouldResemble, emptyTopic)
			})
		})
	})

	Convey("Given root topics exist but no data topics found", t, func() {
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

		Convey("When UpdateDataTopicsPrivate is called", func() {
			respTopics := UpdateDataTopics(ctx, serviceAuthToken, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopics, ShouldResemble, emptyTopic)
			})
		})
	})
}
