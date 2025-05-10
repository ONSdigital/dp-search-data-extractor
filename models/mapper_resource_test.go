package models

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testReleaseDate        = "2010-02-01T00:00:00Z"
	testTitle              = "A Release with A Change Notice for a Future Date"
	testSummary            = "latest summary"
	testCanonicalTopic     = "latest canonical topic"
	testTopics             = []string{"4972", "1245"}
	testURI                = "/peoplepopulationandcommunity/healthandsocialcare/drugusealcoholandsmoking"
	testURIOld             = "an/old/uri"
	testEdition            = "latest edition"
	testContentTypeRelease = "release"
	testCDID               = "A321B"
	testDatasetID          = "ASELECTIONOFNUMBERSANDLETTERS456"
	testMetaDescription    = "latest description"
	testDateChanges        = []ReleaseDateDetails{
		{
			ChangeNotice: "a change_notice",
			Date:         "2059-07-01T12:07:20Z",
		},
	}
	testProvisionalDate = "March 1991"
	testLanguage        = "latest language"
	testSurvey          = "latest survey"
)

func TestMapResourceToSearchDataImport(t *testing.T) {
	Convey("Given a search content update resource", t, func() {
		searchContentUpdate := SearchContentUpdate{
			CanonicalTopic:  testCanonicalTopic,
			CDID:            testCDID,
			ContentType:     testContentTypeRelease,
			DatasetID:       testDatasetID,
			Edition:         testEdition,
			Language:        testLanguage,
			MetaDescription: testMetaDescription,
			ReleaseDate:     testReleaseDate,
			Summary:         testSummary,
			Survey:          testSurvey,
			Title:           testTitle,
			Topics:          testTopics,
			URI:             testURI,
			URIOld:          testURIOld,
			// These fields are only used for ContentType=release
			Cancelled:       false,
			DateChanges:     testDateChanges,
			Finalised:       true,
			ProvisionalDate: testProvisionalDate,
			Published:       false,
		}

		Convey("Then all expected fields are mapped to a SearchDataImport model", func() {
			searchDataImport := MapResourceToSearchDataImport(searchContentUpdate)
			So(searchDataImport, ShouldResemble, SearchDataImport{
				UID:             testURI,
				URI:             testURI,
				Edition:         testEdition,
				DataType:        testContentTypeRelease,
				CDID:            testCDID,
				DatasetID:       testDatasetID,
				MetaDescription: testMetaDescription,
				ReleaseDate:     testReleaseDate,
				Summary:         testSummary,
				Title:           testTitle,
				Topics:          testTopics,
				DateChanges:     testDateChanges,
				Cancelled:       false,
				Finalised:       true,
				ProvisionalDate: testProvisionalDate,
				CanonicalTopic:  testCanonicalTopic,
				Published:       false,
				Language:        testLanguage,
				Survey:          testSurvey,
			})
		})
	})
}
