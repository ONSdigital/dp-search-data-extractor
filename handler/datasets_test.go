package handler

import (
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPopulateCantabularFields(t *testing.T) {
	Convey("Given a dataset Metadata without is_based_on field", t, func() {
		metadata := dataset.Metadata{}

		Convey("When populateCantabularFields is successfully called with a valid datasetDetails", func() {
			dd := &models.DatasetDetails{
				Summary: "This is a test",
			}
			err := populateCantabularFields(ctx, metadata, dd)
			So(err, ShouldBeNil)

			Convey("Then the provided dataset details are not modified", func() {
				So(*dd, ShouldResemble, models.DatasetDetails{
					Summary: "This is a test",
				})
			})
		})

		Convey("When populateCantabularFields is called with a nil datasetDetails, then an error is returned", func() {
			err := populateCantabularFields(ctx, metadata, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "nil DatasetDetails provided")
		})
	})

	Convey("Given a dataset Metadata with is_based_on field, but a non-cantabular type", t, func() {
		metadata := dataset.Metadata{
			DatasetDetails: dataset.DatasetDetails{
				IsBasedOn: &dataset.IsBasedOn{
					Type: "non-cantabular",
				},
			},
		}

		Convey("When populateCantabularFields is successfully called with a valid datasetDetails", func() {
			dd := &models.DatasetDetails{
				Summary: "This is a test",
			}
			err := populateCantabularFields(ctx, metadata, dd)
			So(err, ShouldBeNil)

			Convey("Then the provided dataset details are not modified", func() {
				So(*dd, ShouldResemble, models.DatasetDetails{
					Summary: "This is a test",
				})
			})
		})
	})

	Convey("Given a dataset Metadata with is_based_on field with a cantabular type and 4 dimensions, one being area type", t, func() {
		areaTypeTrue := true
		areaTypeFalse := false
		metadata := dataset.Metadata{
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

		Convey("When populateCantabularFields is successfully called with a valid datasetDetails", func() {
			dd := &models.DatasetDetails{
				Summary: "This is a test",
			}
			err := populateCantabularFields(ctx, metadata, dd)
			So(err, ShouldBeNil)

			Convey("Then the only the non-area-type dimensions are populated in the datasetDetails with the expected values", func() {
				So(*dd, ShouldResemble, models.DatasetDetails{
					Summary: "This is a test",
					Type:    "cantabular_flexible_table",
					Dimensions: []models.Dimension{
						{Name: "dim1", RawLabel: "label 1 (10 categories)", Label: "label 1"},
						{Name: "dim2", RawLabel: "label 2 (12 Categories)", Label: "label 2"},
						{Name: "dim4", RawLabel: "label 4 (1 category)", Label: "label 4"},
					},
				})
			})
		})
	})

	Convey("Given a dataset Metadata with is_based_on field with a cantabular type and a valid population type", t, func() {
		metadata := dataset.Metadata{
			DatasetDetails: dataset.DatasetDetails{
				IsBasedOn: &dataset.IsBasedOn{
					ID:   "UR_HH",
					Type: "cantabular_flexible_table",
				},
			},
		}

		Convey("When populateCantabularFields is successfully called with a valid datasetDetails", func() {
			dd := &models.DatasetDetails{
				Summary: "This is a test",
			}
			err := populateCantabularFields(ctx, metadata, dd)
			So(err, ShouldBeNil)

			Convey("Then the expected population type fields are populated", func() {
				So(*dd, ShouldResemble, models.DatasetDetails{
					Summary:    "This is a test",
					Type:       "cantabular_flexible_table",
					Dimensions: []models.Dimension{},
					PopulationType: models.PopulationType{
						Name:  "UR_HH",
						Label: "All usual residents in households",
					},
				})
			})
		})
	})
}
