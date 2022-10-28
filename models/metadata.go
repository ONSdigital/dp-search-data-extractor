package models

// CMDData provides model for datasetAPI metadata response
type CMDData struct {
	UID            string         `json:"uid"`
	URI            string         `json:"uri"`
	VersionDetails VersionDetails `json:"version"`
	DatasetDetails DatasetDetails `json:"datasetdetails"`
}

// Version represents a version for an edition within a dataset
type VersionDetails struct {
	ReleaseDate string `json:"release_date,omitempty"`
}

// DatasetDetails represents a DatasetDetails for an edition within a dataset
type DatasetDetails struct {
	CanonicalTopic string   `json:"canonical_topic,omitempty"`
	DatasetID      string   `json:"dataset_id,omitempty"`
	Description    string   `json:"description"`
	Edition        string   `json:"edition,omitempty"`
	Keywords       []string `json:"keywords,omitempty"`
	Subtopics      []string `json:"subtopics,omitempty"`
	Title          string   `json:"title"`
	URI            string   `json:"uri,omitempty"`
}
