package models

// ContentPublished provides an avro structure for a Content Published event
type ContentPublished struct {
	URL          string `avro:"url"`
	DataType     string `avro:"data_type"`
	CollectionID string `avro:"collection_id"`
}

// SearchDataImport provides event data for a search data import
type SearchDataImport struct {
	DataType        string `avro:"data_type"`
	JobID           string `avro:"job_id"`
	SearchIndex     string `avro:"search_index"`
	CDID            string `avro:"cdid"`
	DatasetID       string `avro:"dataset_id"`
	Keywords        string `avro:"keywords"`
	MetaDescription string `avro:"meta_description"`
	Summary         string `avro:"summary"`
	ReleaseDate     string `avro:"release_date"`
	Title           string `avro:"title"`
	TraceID         string `avro:"trace_id"`
}
