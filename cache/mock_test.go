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

			topic, err := cacheList.Topic.GetData(ctx, topicSlug)
			So(topic, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetMockDataTopic(t *testing.T) {
	t.Parallel()

	id := "6734"
	slug := topicSlug
	title := "Economy"
	parentID := ""

	Convey("When GetMockTopic is called", t, func() {
		mockTopic := GetMockTopic(id, slug, title, parentID)

		Convey("Then the mock topic is returned", func() {
			So(mockTopic, ShouldNotBeNil)
			So(mockTopic.ID, ShouldEqual, id)
			So(mockTopic.Slug, ShouldEqual, slug)
			So(mockTopic.LocaliseKeyName, ShouldEqual, title)
			So(mockTopic.ParentID, ShouldNotBeNil)
		})
	})
}
