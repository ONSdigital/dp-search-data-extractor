package models_test

import (
	"testing"

	"github.com/ONSdigital/dp-search-data-extractor/models"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	someDataType  = "datatype"
	someCDID      = "CDID"
	someDatasetID = "datasetID"
	someEdition   = "edition"

	somekeyword0 = "keyword0"
	somekeyword1 = "keyword1"
	somekeyword2 = "keyword2"
	somekeyword3 = "keyword3"

	someMetaDescription = "meta desc"
	someReleaseDate     = "2021-12-13"
	someSummary         = "Some Amazing Summary"
	someTitle           = "Some Incredible Title"

	sometopic0 = "topic0"
	sometopic1 = "topic1"
)

func TestMapZebedeeDataToSearchDataImport(t *testing.T) {
	Convey("Given some valid zebedee data with  ", t, func() {
		zebendeeData := models.ZebedeeData{
			UID:      someTitle,
			DataType: someDataType,
			Description: models.Description{
				CDID:            someCDID,
				DatasetID:       someDatasetID,
				Edition:         someEdition,
				Keywords:        []string{somekeyword0, somekeyword1, somekeyword2, somekeyword3},
				MetaDescription: someMetaDescription,
				ReleaseDate:     someReleaseDate,
				Summary:         someSummary,
				Title:           someTitle,
				Topics:          []string{sometopic0, sometopic1},
			},
		}
		Convey("When mapped with a default keywords limit", func() {
			result := models.MapZebedeeDataToSearchDataImport(zebendeeData, -1)
			Convey("Then the result should be validly mapped with 4 keywords", func() {
				So(result.UID, ShouldResemble, someTitle)
				So(result.DataType, ShouldResemble, someDataType)
				So(result.CDID, ShouldResemble, someCDID)
				So(result.DatasetID, ShouldResemble, someDatasetID)
				So(result.MetaDescription, ShouldResemble, someMetaDescription)
				So(result.ReleaseDate, ShouldResemble, someReleaseDate)
				So(result.Summary, ShouldResemble, someSummary)
				So(result.Title, ShouldResemble, someTitle)

				So(result.Keywords, ShouldNotBeEmpty)
				So(result.Keywords, ShouldHaveLength, 4)
				So(result.Keywords[0], ShouldResemble, somekeyword0)
				So(result.Keywords[1], ShouldResemble, somekeyword1)
				So(result.Keywords[2], ShouldResemble, somekeyword2)
				So(result.Keywords[3], ShouldResemble, somekeyword3)

				So(result.Topics, ShouldNotBeEmpty)
				So(result.Topics, ShouldHaveLength, 2)
				So(result.Topics[0], ShouldResemble, sometopic0)
				So(result.Topics[1], ShouldResemble, sometopic1)
			})
		})
		Convey("When mapped with a keywords limit of 2", func() {
			result := models.MapZebedeeDataToSearchDataImport(zebendeeData, 2)
			Convey("Then the result should be validly mapped with 2 keywords", func() {
				So(result.DataType, ShouldResemble, someDataType)
				So(result.CDID, ShouldResemble, someCDID)
				So(result.DatasetID, ShouldResemble, someDatasetID)
				So(result.MetaDescription, ShouldResemble, someMetaDescription)
				So(result.ReleaseDate, ShouldResemble, someReleaseDate)
				So(result.Summary, ShouldResemble, someSummary)
				So(result.Title, ShouldResemble, someTitle)

				So(result.Keywords, ShouldNotBeEmpty)
				So(result.Keywords, ShouldHaveLength, 2)
				So(result.Keywords[0], ShouldResemble, somekeyword0)
				So(result.Keywords[1], ShouldResemble, somekeyword1)
			})
		})
	})
}

func TestRectifyKeywords_WithEmptyKeywordsAndDefaultLimit(t *testing.T) {
	Convey("Given an empty keywords as received from zebedee  with default keywords limit", t, func() {
		testKeywords := []string{""}
		Convey("When passed to rectify the keywords with default keywords limit", func() {
			actual := models.RectifyKeywords(testKeywords, -1)
			Convey("Then keywords should be rectified as expected empty keywords", func() {
				So(actual, ShouldResemble, testKeywords)
			})
		})
	})
}

func TestRectifyKeywords_WithTrimmingKeywordsAndDefaultLimit(t *testing.T) {
	Convey("Given un-trimmed keywords as received from zebedee  with default keywords limit", t, func() {
		testKeywords := []string{"  testKeywords1,testKeywords2   "}
		Convey("When passed to rectify keywords with keywords limits as 2", func() {
			actual := models.RectifyKeywords(testKeywords, -1)
			Convey("Then keywords should be rectified with correct size and expected trimmed elements", func() {
				expectedKeywords := []string{"testKeywords1", "testKeywords2"}
				So(actual, ShouldResemble, expectedKeywords)
			})
		})
	})
}

func TestRectifyKeywords_LessFourKeywordsAndDefaultLimit(t *testing.T) {
	Convey("Given a keywords as received from zebedee with default keywords limit", t, func() {
		testKeywords := []string{"testkeyword1,testkeyword2", "testkeyword3,testKeywords4"}
		Convey("When passed to rectify the keywords with default keywords limit", func() {
			actual := models.RectifyKeywords(testKeywords, -1)
			Convey("Then all the same keywords should be returned", func() {
				expectedKeywords := []string{"testkeyword1", "testkeyword2", "testkeyword3", "testKeywords4"}
				So(actual, ShouldResemble, expectedKeywords)
			})
		})
	})
}

func TestRectifyKeywords_WithEightKeywordsAndZeroAsLimit(t *testing.T) {
	Convey("Given a keywords as received from zebedee with zero keywords limit", t, func() {
		testKeywords := []string{"testkeyword1,testkeyword2", "testkeyword3,testKeywords4", "testkeyword5,testKeywords6,testkeyword7,testKeywords8"}
		Convey("When passed to rectify the keywords with default keywords limits", func() {
			actual := models.RectifyKeywords(testKeywords, 0)
			Convey("Then keywords should be rectified with empty keyword elements", func() {
				expectedKeywords := []string{""}
				So(actual, ShouldResemble, expectedKeywords)
			})
		})
	})
}

func TestRectifyKeywords_WithEightKeywordsAndDefaultLimit(t *testing.T) {
	Convey("Given a keywords as received from zebedee with default keywords limit", t, func() {
		testKeywords := []string{"testkeyword1,testkeyword2", "testkeyword3,testKeywords4", "testkeyword5,testKeywords6,testkeyword7,testKeywords8"}
		Convey("When passed to rectify the keywords with default keywords limits", func() {
			actual := models.RectifyKeywords(testKeywords, -1)
			Convey("Then keywords should be rectified with correct size with all keyword elements", func() {
				expectedKeywords := []string{"testkeyword1", "testkeyword2", "testkeyword3", "testKeywords4", "testkeyword5", "testKeywords6", "testkeyword7", "testKeywords8"}
				So(actual, ShouldResemble, expectedKeywords)
			})
		})
	})
}

func TestRectifyKeywords_EightKeywordsAndFiveAsLimit(t *testing.T) {
	Convey("Given a keywords as received from zebedee with five keywords limit", t, func() {
		testKeywords := []string{"testkeyword1,testkeyword2", "testkeyword3,testKeywords4", "testkeyword5,testKeywords6,testkeyword7,testKeywords8"}
		Convey("When passed to rectify the keywords with keywords limit as 5", func() {
			actual := models.RectifyKeywords(testKeywords, 5)
			Convey("Then keywords should be rectified with correct size with expected elements", func() {
				expectedKeywords := []string{"testkeyword1", "testkeyword2", "testkeyword3", "testKeywords4", "testkeyword5"}
				So(actual, ShouldResemble, expectedKeywords)
			})
		})
	})
}

func TestRectifyKeywords_EightKeywordsAndTenAsLimit(t *testing.T) {
	Convey("Given a keywords as received from zebedee with ten keywords limit", t, func() {
		testKeywords := []string{"testkeyword1,testkeyword2", "testkeyword3,testKeywords4", "testkeyword5,testKeywords6,testkeyword7,testKeywords8"}
		Convey("When passed to rectify the keywords with keywords limit as 5", func() {
			actual := models.RectifyKeywords(testKeywords, 10)
			Convey("Then keywords should be rectified with correct size with expected elements", func() {
				expectedKeywords := []string{"testkeyword1", "testkeyword2", "testkeyword3", "testKeywords4", "testkeyword5", "testKeywords6", "testkeyword7", "testKeywords8"}
				So(actual, ShouldResemble, expectedKeywords)
			})
		})
	})
}

func TestValidate_WithNiltopics(t *testing.T) {
	Convey("Given a nil topics array as received from zebedee", t, func() {
		Convey("When passed to validate the topics", func() {
			actual := models.ValidateTopics(nil)
			Convey("Then topics should be populated with an empty string array", func() {
				So(actual, ShouldResemble, []string{""})
			})
		})
	})
}

func TestMapDatasetVersionMetadataToSearchDataImport(t *testing.T) {
	Convey("Given some valid DatasetAPI data with", t, func() {
		CMDTestData := models.CMDData{
			UID: "someuid",
			VersionDetails: models.VersionDetails{
				ReleaseDate: someReleaseDate,
			},
			DatasetDetails: models.DatasetDetails{
				Title:       someTitle,
				Description: someMetaDescription,
				Keywords:    []string{somekeyword0, somekeyword1, somekeyword2, somekeyword3},
			},
		}
		Convey("When passed to rectify the keywords with keywords limit as 5", func() {
			actual := models.MapVersionMetadataToSearchDataImport(CMDTestData)

			Convey("Then keywords should be rectified with correct size with expected elements", func() {
				So(actual.UID, ShouldResemble, "someuid")
				So(actual.ReleaseDate, ShouldResemble, someReleaseDate)
				So(actual.Title, ShouldResemble, someTitle)
				So(actual.Keywords, ShouldNotBeEmpty)
				So(actual.Keywords, ShouldHaveLength, 4)
				So(actual.Keywords[0], ShouldResemble, somekeyword0)
				So(actual.Keywords[1], ShouldResemble, somekeyword1)
				So(actual.Keywords[2], ShouldResemble, somekeyword2)
				So(actual.Keywords[3], ShouldResemble, somekeyword3)
			})
		})
	})
}
