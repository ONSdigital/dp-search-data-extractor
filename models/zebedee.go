package models

// ZebedeeData provides model for zebedee publisheddata response
type ZebedeeData struct {
	Uid            string         `json:"uid"`
	DataType    string      `json:"type"`
	Description Description `json:"description"`
}

type Description struct {
	CDID            string   `json:"cdid"`
	DatasetID       string   `json:"datasetId"`
	Edition         string   `json:"edition"`
	Keywords        []string `json:"keywords,omitempty"`
	MetaDescription string   `json:"metaDescription"`
	ReleaseDate     string   `json:"releaseDate"`
	Summary         string   `json:"summary"`
	Title           string   `json:"title"`
}
