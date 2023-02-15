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

	Convey("Given a dataset Metadata with is_based_on field with a cantabular type and 3 dimensions", t, func() {
		metadata := dataset.Metadata{
			DatasetDetails: dataset.DatasetDetails{
				IsBasedOn: &dataset.IsBasedOn{
					Type: "cantabular_flexible_table",
				},
			},
			Version: dataset.Version{
				Dimensions: []dataset.VersionDimension{
					{ID: "dim1", Label: "label 1 (10 categories)"},
					{ID: "dim2", Label: "label 2 (12 Categories)"},
					{ID: "dim3", Label: "label 3 (1 category)"},
				},
			},
		}

		Convey("When populateCantabularFields is successfully called with a nil datasetDetails", func() {
			dd := &models.DatasetDetails{
				Summary: "This is a test",
			}
			err := populateCantabularFields(ctx, metadata, dd)
			So(err, ShouldBeNil)

			Convey("Then it is initialised", func() {
				So(*dd, ShouldResemble, models.DatasetDetails{
					Summary: "This is a test",
					Type:    "cantabular_flexible_table",
					Dimensions: []models.Dimension{
						{Name: "dim1", RawLabel: "label 1 (10 categories)", Label: "label 1"},
						{Name: "dim2", RawLabel: "label 2 (12 Categories)", Label: "label 2"},
						{Name: "dim3", RawLabel: "label 3 (1 category)", Label: "label 3"},
					},
				})
			})
		})
	})
}
