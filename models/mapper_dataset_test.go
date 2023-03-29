package models_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testReleaseDate    = "testReleaseDate"
	testTitle          = "testTitle"
	testSummary        = "testSummary"
	testCanonicalTopic = "testCanonicalTopic"
	testTopics         = []string{"t1", "t2", "t3"}
	testKeywords       = []string{"k1", "k2", "k3"}
	testPath           = "/test/url/path"
	testURL            = fmt.Sprintf("http://testhost:1234%s", testPath)
)

var ctx = context.Background()

func TestMapDatasetMetadataValues(t *testing.T) {
	Convey("Given some valid dataset api metadata", t, func() {
		metadata := &dataset.Metadata{
			Version: dataset.Version{
				ReleaseDate: testReleaseDate,
			},
			DatasetDetails: dataset.DatasetDetails{
				Title:          testTitle,
				Description:    testSummary,
				CanonicalTopic: testCanonicalTopic,
				Subtopics:      testTopics,
				Keywords:       &testKeywords,
			},
			DatasetLinks: dataset.Links{
				LatestVersion: dataset.Link{
					URL: testURL,
				},
			},
		}

		Convey("Then all expected fields are mapped to a SearchDataImport model", func() {
			s := models.SearchDataImport{}
			err := s.MapDatasetMetadataValues(ctx, metadata)
			So(err, ShouldBeNil)
			So(s, ShouldResemble, models.SearchDataImport{
				ReleaseDate:    testReleaseDate,
				Title:          testTitle,
				Summary:        testSummary,
				CanonicalTopic: testCanonicalTopic,
				Topics:         testTopics,
				Keywords:       testKeywords,
				URI:            testPath,
			})
		})
	})

	Convey("trying to map a nil metadata value returns the expected error", t, func() {
		s := models.SearchDataImport{}
		err := s.MapDatasetMetadataValues(ctx, nil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "nil metadata cannot be mapped")

		Convey("And the search data import event is not modified", func() {
			So(s, ShouldResemble, models.SearchDataImport{})
		})
	})

	Convey("Given a dataset api metadata with a malformed URL value", t, func() {
		metadata := &dataset.Metadata{
			DatasetLinks: dataset.Links{
				LatestVersion: dataset.Link{
					URL: "wrong£%$@",
				},
			},
		}

		Convey("Then trying to map the values to a search data import event fails with the expected error", func() {
			s := models.SearchDataImport{}
			err := s.MapDatasetMetadataValues(ctx, metadata)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "error parsing the metadata uri: parse \"wrong£%$@\": invalid URL escape \"%$@\"")

			Convey("And the search data import event is not modified", func() {
				So(s, ShouldResemble, models.SearchDataImport{})
			})
		})
	})
}

func TestPopulateCantabularFields(t *testing.T) {
	Convey("Given a dataset Metadata without is_based_on field", t, func() {
		metadata := &dataset.Metadata{}

		Convey("When PopulateCantabularFields is called on a valid search data import struct", func() {
			s := &models.SearchDataImport{
				Summary: testSummary,
			}
			s.PopulateCantabularFields(ctx, metadata)

			Convey("Then the search data import is not modified", func() {
				So(*s, ShouldResemble, models.SearchDataImport{
					Summary: testSummary,
				})
			})
		})
	})

	Convey("Given a dataset Metadata with is_based_on field, but a non-cantabular type", t, func() {
		metadata := &dataset.Metadata{
			DatasetDetails: dataset.DatasetDetails{
				IsBasedOn: &dataset.IsBasedOn{
					Type: "non-cantabular",
				},
			},
		}

		Convey("When PopulateCantabularFields is called on a valid search data import struct", func() {
			s := &models.SearchDataImport{
				Summary: testSummary,
			}
			s.PopulateCantabularFields(ctx, metadata)

			Convey("Then the search data import is not modified", func() {
				So(*s, ShouldResemble, models.SearchDataImport{
					Summary: testSummary,
				})
			})
		})
	})

	Convey("Given a dataset metadata with is_based_on field with a cantabular type and 4 dimensions, one being area type", t, func() {
		areaTypeTrue := true
		areaTypeFalse := false
		metadata := &dataset.Metadata{
			DatasetDetails: dataset.DatasetDetails{
				IsBasedOn: &dataset.IsBasedOn{
					Type: "cantabular_flexible_table",
				},
			},
			Version: dataset.Version{
				Dimensions: []dataset.VersionDimension{
					{ID: "dim1", Label: "label 1 (10 categories)"},
					{ID: "dim2", Label: "label 2 (12 Categories)", IsAreaType: &areaTypeFalse},
					{ID: "dim3", IsAreaType: &areaTypeTrue},
					{ID: "dim4", Label: "label 4 (1 category)"},
				},
			},
		}

		Convey("When PopulateCantabularFields is called on a valid search data import struct", func() {
			s := &models.SearchDataImport{
				Summary:  testSummary,
				DataType: "dataset_landing_page",
			}
			s.PopulateCantabularFields(ctx, metadata)

			Convey("Then only the non-area-type dimensions are populated, with the expected values", func() {
				So(*s, ShouldResemble, models.SearchDataImport{
					Summary:  testSummary,
					DataType: "dataset_landing_page",
					Dimensions: []models.Dimension{
						{Name: "dim1", RawLabel: "label 1 (10 categories)", Label: "label 1", AggKey: "dim1###label 1"},
						{Name: "dim2", RawLabel: "label 2 (12 Categories)", Label: "label 2", AggKey: "dim2###label 2"},
						{Name: "dim4", RawLabel: "label 4 (1 category)", Label: "label 4", AggKey: "dim4###label 4"},
					},
				})
			})
		})
	})

	Convey("Given a dataset metadata with is_based_on field with a cantabular type and a valid population type", t, func() {
		metadata := &dataset.Metadata{
			DatasetDetails: dataset.DatasetDetails{
				IsBasedOn: &dataset.IsBasedOn{
					ID:   "UR_HH",
					Type: "cantabular_flexible_table",
				},
			},
		}

		Convey("When PopulateCantabularFields is called on a valid search data import struct", func() {
			s := &models.SearchDataImport{
				Summary:  testSummary,
				DataType: "dataset_landing_page",
			}
			s.PopulateCantabularFields(ctx, metadata)

			Convey("Then the expected population type fields are populated", func() {
				So(*s, ShouldResemble, models.SearchDataImport{
					Summary:    testSummary,
					DataType:   "dataset_landing_page",
					Dimensions: []models.Dimension{},
					PopulationType: models.PopulationType{
						Name:   "UR_HH",
						Label:  "All usual residents in households",
						AggKey: "UR_HH###All usual residents in households",
					},
				})
			})
		})
	})
}

func TestGetURI(t *testing.T) {
	dl := dataset.Links{
		LatestVersion: dataset.Link{
			URL: "dataset_link",
		},
	}

	dd := dataset.DatasetDetails{
		Links: dataset.Links{
			Version: dataset.Link{
				URL: "dataset_details_link",
			},
		},
	}

	v := dataset.Version{
		Links: dataset.Links{
			Version: dataset.Link{
				URL: "version_link",
			},
		},
	}

	Convey("Given a dataset metadata struct with dataset links, dataset details links and version links", t, func() {
		dm := &dataset.Metadata{
			DatasetLinks:   dl,
			DatasetDetails: dd,
			Version:        v,
		}

		Convey("Then GetURI should return the URL under dataset links latest version", func() {
			So(models.GetURI(dm), ShouldEqual, "dataset_link")
		})
	})

	Convey("Given a dataset metadata struct with dataset details links and version links", t, func() {
		dm := &dataset.Metadata{
			DatasetDetails: dd,
			Version:        v,
		}

		Convey("Then GetURI should return the URL under dataset details version link", func() {
			So(models.GetURI(dm), ShouldEqual, "dataset_details_link")
		})
	})

	Convey("Given a dataset metadata struct with only version links", t, func() {
		dm := &dataset.Metadata{
			Version: v,
		}

		Convey("Then GetURI should return the version link", func() {
			So(models.GetURI(dm), ShouldEqual, "version_link")
		})
	})

	Convey("GetURI with a nil metadata structure returns an empty string", t, func() {
		So(models.GetURI(nil), ShouldEqual, "")
	})
}
