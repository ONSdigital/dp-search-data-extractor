package models_test

import (
	"testing"

	"github.com/ONSdigital/dp-search-data-extractor/models"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	someRelease          = "release"
	someCDID             = "CDID"
	someChangeNotice     = "Delay on publication"
	someChangeNoticeDate = "2021-12-14"
	someDatasetID        = "datasetID"
	someEdition          = "edition"

	somekeyword0 = "keyword0"
	somekeyword1 = "keyword1"
	somekeyword2 = "keyword2"
	somekeyword3 = "keyword3"

	someMetaDescription = "meta desc"
	someProvisionalDate = "2021-12-12"
	someReleaseDate     = "2021-12-13"
	someSummary         = "Some Amazing Summary"
	someTitle           = "Some Incredible Title"
	someSurvey          = "Some survey"
	someLanguage        = "some language"
	canonicalTopic      = "some topic"
	someURI             = "/economy/uri/for/testing"

	sometopic0 = "topic0"
	sometopic1 = "topic1"
)

func TestMapZebedeeDataToSearchDataImport(t *testing.T) {
	Convey("Given some valid zebedee data", t, func() {
		zebedeeData := models.ZebedeeData{
			UID:      someTitle,
			URI:      someURI,
			DataType: someRelease,
			DateChanges: []models.ReleaseDateChange{
				{
					ChangeNotice: someChangeNotice,
					Date:         someChangeNoticeDate,
				},
			},
			Description: models.Description{
				Cancelled:       false,
				CDID:            someCDID,
				DatasetID:       someDatasetID,
				Finalised:       true,
				Keywords:        []string{somekeyword0, somekeyword1, somekeyword2, somekeyword3},
				MetaDescription: someMetaDescription,
				ProvisionalDate: someProvisionalDate,
				Published:       true,
				ReleaseDate:     someReleaseDate,
				Summary:         someSummary,
				Title:           someTitle,
				Topics:          []string{sometopic0, sometopic1},
				Survey:          someSurvey,
				Language:        someLanguage,
				CanonicalTopic:  canonicalTopic,
			},
		}
		zebedeeDataWithEdition := models.ZebedeeData{
			UID:      someTitle,
			URI:      someURI,
			DataType: someRelease,
			DateChanges: []models.ReleaseDateChange{
				{
					ChangeNotice: someChangeNotice,
					Date:         someChangeNoticeDate,
				},
			},
			Description: models.Description{
				Cancelled:       false,
				CDID:            someCDID,
				DatasetID:       someDatasetID,
				Edition:         someEdition,
				Finalised:       true,
				Keywords:        []string{somekeyword0, somekeyword1, somekeyword2, somekeyword3},
				MetaDescription: someMetaDescription,
				ProvisionalDate: someProvisionalDate,
				Published:       true,
				ReleaseDate:     someReleaseDate,
				Summary:         someSummary,
				Title:           someTitle,
				Topics:          []string{sometopic0, sometopic1},
				Survey:          someSurvey,
				Language:        someLanguage,
				CanonicalTopic:  canonicalTopic,
			},
		}
		Convey("And mapped with a default keywords limit when edition is present", func() {
			result := models.MapZebedeeDataToSearchDataImport(zebedeeDataWithEdition, -1)
			Convey("Then the result should be validly mapped with 4 keywords when edition is present", func() {
				So(result.UID, ShouldResemble, someURI)
				So(result.URI, ShouldResemble, someURI)
				So(result.Edition, ShouldResemble, someEdition)
				So(result.DataType, ShouldResemble, someRelease)
				So(result.CDID, ShouldResemble, someCDID)
				So(result.DatasetID, ShouldResemble, someDatasetID)
				So(result.MetaDescription, ShouldResemble, someMetaDescription)
				So(result.ReleaseDate, ShouldResemble, someReleaseDate)
				So(result.Summary, ShouldResemble, someSummary)
				So(result.Title, ShouldResemble, someTitle)
				So(result.Survey, ShouldResemble, someSurvey)
				So(result.Language, ShouldResemble, someLanguage)
				So(result.CanonicalTopic, ShouldResemble, canonicalTopic)

				So(result.DateChanges, ShouldHaveLength, 1)
				So(result.DateChanges[0].ChangeNotice, ShouldEqual, someChangeNotice)
				So(result.DateChanges[0].Date, ShouldEqual, someChangeNoticeDate)

				So(result.Cancelled, ShouldBeFalse)
				So(result.Finalised, ShouldBeTrue)
				So(result.Published, ShouldBeTrue)

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
		Convey("And mapped with a default keywords limit", func() {
			result := models.MapZebedeeDataToSearchDataImport(zebedeeData, -1)
			Convey("Then the result should be validly mapped with 4 keywords", func() {
				So(result.UID, ShouldResemble, someURI)
				So(result.URI, ShouldResemble, someURI)
				So(result.DataType, ShouldResemble, someRelease)
				So(result.CDID, ShouldResemble, someCDID)
				So(result.DatasetID, ShouldResemble, someDatasetID)
				So(result.MetaDescription, ShouldResemble, someMetaDescription)
				So(result.ReleaseDate, ShouldResemble, someReleaseDate)
				So(result.Summary, ShouldResemble, someSummary)
				So(result.Title, ShouldResemble, someTitle)
				So(result.Survey, ShouldResemble, someSurvey)
				So(result.Language, ShouldResemble, someLanguage)
				So(result.CanonicalTopic, ShouldResemble, canonicalTopic)

				So(result.DateChanges, ShouldHaveLength, 1)
				So(result.DateChanges[0].ChangeNotice, ShouldEqual, someChangeNotice)
				So(result.DateChanges[0].Date, ShouldEqual, someChangeNoticeDate)

				So(result.Cancelled, ShouldBeFalse)
				So(result.Finalised, ShouldBeTrue)
				So(result.Published, ShouldBeTrue)

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
		Convey("And mapped with a keywords limit of 2", func() {
			result := models.MapZebedeeDataToSearchDataImport(zebedeeData, 2)
			Convey("Then the result should be validly mapped with 2 keywords", func() {
				So(result.DataType, ShouldResemble, someRelease)
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
				expectedKeywords := []string{}
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
	Convey("Given a topics array is missing in zebedee response", t, func() {
		Convey("When passed to validate the topics", func() {
			actual := models.ValidateTopics(nil)
			Convey("Then topics should be populated with an empty string array", func() {
				So(actual, ShouldResemble, []string{""})
			})
		})
	})
}

func TestValidate_WithEmptytopics(t *testing.T) {
	Convey("Given a empty topics array as received from zebedee", t, func() {
		Convey("When passed to validate the topics", func() {
			actual := models.ValidateTopics([]string{""})
			Convey("Then topics should be populated with an empty string array", func() {
				So(actual, ShouldResemble, []string{""})
			})
		})
	})
}

func TestValidate_WithEmptytopicsSize0(t *testing.T) {
	Convey("Given a topics array of size 0 received from zebedee", t, func() {
		Convey("When passed to validate the topics", func() {
			actual := models.ValidateTopics([]string{})
			Convey("Then topics should be populated with an string array of size 0", func() {
				So(actual, ShouldResemble, []string{})
			})
		})
	})
}
