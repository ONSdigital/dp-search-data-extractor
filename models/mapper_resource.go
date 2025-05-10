package models

// MapResourceToSearchDataImport Performs default mapping of a SearchContentUpdate struct to a SearchDataImport struct.
func MapResourceToSearchDataImport(resource SearchContentUpdate) SearchDataImport {
	searchDataImport := SearchDataImport{
		UID:             resource.URI,
		URI:             resource.URI,
		Title:           resource.Title,
		DataType:        resource.ContentType,
		Summary:         resource.Summary,
		Survey:          resource.Survey,
		MetaDescription: resource.MetaDescription,
		ReleaseDate:     resource.ReleaseDate,
		Language:        resource.Language,
		Edition:         resource.Edition,
		DatasetID:       resource.DatasetID,
		CDID:            resource.CDID,
		CanonicalTopic:  resource.CanonicalTopic,
		Topics:          []string{},
	}

	// Assign topics only if they're not nil
	if resource.Topics != nil {
		searchDataImport.Topics = resource.Topics
	}

	if resource.ContentType == ReleaseDataType {
		searchDataImport.Cancelled = resource.Cancelled
		searchDataImport.Finalised = resource.Finalised
		searchDataImport.Published = resource.Published
		searchDataImport.ProvisionalDate = resource.ProvisionalDate
		searchDataImport.DateChanges = resource.DateChanges
	}
	return searchDataImport
}
