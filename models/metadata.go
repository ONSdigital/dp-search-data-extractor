package models

// CMDData provides model for datasetAPI metadata response
type CMDData struct {
	Uid            string         `json:"uid"`
	VersionDetails VersionDetails `json:"version"`
	DatasetDetails DatasetDetails `json:"datasetdetails"`
}

// Version represents a version for an edition within a dataset
type VersionDetails struct {
	ReleaseDate string `json:"release_date,omitempty"`
}

// DatasetDetails represents a DatasetDetails for an edition within a dataset
type DatasetDetails struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords,omitempty"`
}
