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
		TraceID:         resource.TraceID,
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

// MapResourceToSearchContentDelete Performs default mapping of a SearchContentUpdate struct to a SearchContentDeleted struct.
func MapResourceToSearchContentDelete(resource SearchContentUpdate) SearchContentDeleted {
	searchContentDeleted := SearchContentDeleted{
		URI:          resource.URIOld,
		CollectionID: resource.CDID,
		TraceID:      resource.TraceID,
	}
	return searchContentDeleted
}
