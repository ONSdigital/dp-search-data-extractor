package models

import (
	"strings"
)

// MapZebedeeDataToSearchDataImport Performs default mapping of zebedee data to a SearchDataImport struct.
// It also optionally takes a limit which truncates the keywords to the desired amount. This value can be -1 for no
// truncation.
func MapZebedeeDataToSearchDataImport(zebedeeData ZebedeeData, keywordsLimit int) SearchDataImport {
	searchData := SearchDataImport{
		DataType:        zebedeeData.DataType,
		CDID:            zebedeeData.Description.CDID,
		DatasetID:       zebedeeData.Description.DatasetID,
		Keywords:        RectifyKeywords(zebedeeData.Description.Keywords, keywordsLimit),
		MetaDescription: zebedeeData.Description.MetaDescription,
		Summary:         zebedeeData.Description.Summary,
		ReleaseDate:     zebedeeData.Description.ReleaseDate,
		Title:           zebedeeData.Description.Title,
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

// MapDatasetVersionToSearchDataImport performs default mapping of datasetAPI data to a version struct.
func MapDatasetVersionToSearchDataImport(versionMetadata VersionMetadata) SearchDataVersionMetadataImport {
	versionData := SearchDataVersionMetadataImport{
		CollectionId: versionMetadata.CollectionId,
		Edition:      versionMetadata.Edition,
		ID:           versionMetadata.ID,
		DatasetId:    versionMetadata.DatasetId,
		Version:      versionMetadata.Version,
		ReleaseDate:  versionMetadata.ReleaseDate,
	}
	return versionData
}
