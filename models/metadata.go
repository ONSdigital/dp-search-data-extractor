package models

// CMDData provides model for datasetAPI metadata response
type CMDData struct {
	UID            string
	URI            string
	VersionDetails VersionDetails
	DatasetDetails DatasetDetails
}

// Version represents a version for an edition within a dataset
type VersionDetails struct {
	ReleaseDate string
}

// DatasetDetails represents a DatasetDetails for an edition within a dataset
type DatasetDetails struct {
	CanonicalTopic string
	DatasetID      string
	Summary        string
	Edition        string
	Keywords       []string
	Subtopics      []string
	Title          string
	URI            string
	Type           string
}
