package cache

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const topicSlug = "economy"

func TestGetMockCacheList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("When GetMockCacheList is called", t, func() {
		cacheList, err := GetMockCacheList(ctx)

		Convey("Then the list of cache should be returned", func() {
			So(cacheList, ShouldNotBeNil)
			So(err, ShouldBeNil)

			So(cacheList.Topic, ShouldNotBeNil)

			topic, err := cacheList.Topic.GetTopic(ctx, topicSlug)
			So(topic, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetMockRootTopic(t *testing.T) {
	t.Parallel()

	rootTopicID := "root-topic-cache"
	slug := "root"

	Convey("When GetMockRootTopic is called", t, func() {
		mockTopic := GetMockRootTopic(rootTopicID)

		Convey("Then the mocked root topic is returned", func() {
			So(mockTopic, ShouldNotBeNil)
			So(mockTopic.ID, ShouldEqual, rootTopicID)
			So(mockTopic.Slug, ShouldEqual, slug)

			subtopic, exists := mockTopic.List.Get("economy")
			So(exists, ShouldBeTrue)
			So(subtopic, ShouldResemble, Subtopic{ID: "6734", Slug: "economy", LocaliseKeyName: "Economy", ReleaseDate: timeHelper("2022-10-10T08:30:00Z"), ParentID: ""})
		})
	})
}
