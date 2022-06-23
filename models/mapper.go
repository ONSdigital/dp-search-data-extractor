package models

import (
	"context"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"
)

const (
	ReleaseDataType = "release"
)

// MapZebedeeDataToSearchDataImport Performs default mapping of zebedee data to a SearchDataImport struct.
// It also optionally takes a limit which truncates the keywords to the desired amount. This value can be -1 for no
// truncation.
func MapZebedeeDataToSearchDataImport(zebedeeData ZebedeeData, keywordsLimit int) SearchDataImport {
	searchData := SearchDataImport{
		UID:             zebedeeData.Description.Title,
		URI:             zebedeeData.URI,
		DataType:        zebedeeData.DataType,
		CDID:            zebedeeData.Description.CDID,
		DatasetID:       zebedeeData.Description.DatasetID,
		Keywords:        RectifyKeywords(zebedeeData.Description.Keywords, keywordsLimit),
		MetaDescription: zebedeeData.Description.MetaDescription,
		Summary:         zebedeeData.Description.Summary,
		ReleaseDate:     zebedeeData.Description.ReleaseDate,
		Title:           zebedeeData.Description.Title,
		Topics:          zebedeeData.Description.Topics,
	}
	if zebedeeData.DataType == ReleaseDataType {
		logData := log.Data{
			"zebedeeRCData": zebedeeData,
		}
		log.Info(context.Background(), "zebedee release calender data", logData)
		for _, data := range zebedeeData.DateChanges {
			searchData.DateChanges = append(searchData.DateChanges, ReleaseDateDetails(data))
		}
		searchData.Published = zebedeeData.Description.Published
		searchData.Cancelled = zebedeeData.Description.Cancelled
		searchData.Finalised = zebedeeData.Description.Finalised
		searchData.ProvisionalDate = zebedeeData.Description.ProvisionalDate
		searchData.Survey = zebedeeData.Description.Survey
		searchData.Language = zebedeeData.Description.Language
		searchData.CanonicalTopic = zebedeeData.Description.CanonicalTopic
	}
	return searchData
}

// RectifyKeywords sanitises a slice of keywords, splitting any that contain commas into seperate keywords and trimming
// any whitespace. It also optionally takes a limit which truncates the keywords to the desired amount. This value can
// be -1 for no truncation.
func RectifyKeywords(keywords []string, keywordsLimit int) []string {
	var strArray []string
	rectifiedKeywords := make([]string, 0)

	if keywordsLimit == 0 {
		return []string{""}
	}

	for i := range keywords {
		strArray = strings.Split(keywords[i], ",")

		for j := range strArray {
			keyword := strings.TrimSpace(strArray[j])
			rectifiedKeywords = append(rectifiedKeywords, keyword)
		}
	}

	if (len(rectifiedKeywords) < keywordsLimit) || (keywordsLimit == -1) {
		return rectifiedKeywords
	}

	return rectifiedKeywords[:keywordsLimit]
}

func ValidateTopics(topics []string) []string {
	if topics == nil {
		topics = []string{""}
	}
	return topics
}

// MapDatasetVersionToSearchDataImport performs default mapping of datasetAPI data to a version metadata struct.
func MapVersionMetadataToSearchDataImport(cmdData CMDData) SearchDataImport {
	versionMetaData := SearchDataImport{
		UID:             cmdData.UID,
		URI:             cmdData.URI,
		ReleaseDate:     cmdData.VersionDetails.ReleaseDate,
		Keywords:        cmdData.DatasetDetails.Keywords,
		MetaDescription: cmdData.DatasetDetails.Description,
		Title:           cmdData.DatasetDetails.Title,
	}

	return versionMetaData
}
