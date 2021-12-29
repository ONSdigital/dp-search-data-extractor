package models

// Version represents a version for an edition within a dataset
type VersionMetadata struct {
	CollectionId string `json:"collection_id,omitempty"`
	Edition      string `json:"edition,omitempty"`
	ID           string `json:"id,omitempty"`
	DatasetId    string `json:"dataset_id,omitempty"`
	ReleaseDate  string `json:"release_date,omitempty"`
	Version      string `json:"version,omitempty"`
}
