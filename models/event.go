package models

// ContentPublished provides an avro structure for a Content Published event
type ContentPublished struct {
	URI          string `avro:"uri"`
	DataType     string `avro:"data_type"`
	CollectionID string `avro:"collection_id"`
	JobID        string `avro:"job_id"`
	SearchIndex  string `avro:"search_index"`
	TraceID      string `avro:"trace_id"`
}

// SearchDataImport provides event data for a search data import
type SearchDataImport struct {
	UID             string   `avro:"uid"`
	DataType        string   `avro:"data_type"`
	JobID           string   `avro:"job_id"`
	SearchIndex     string   `avro:"search_index"`
	CDID            string   `avro:"cdid"`
	DatasetID       string   `avro:"dataset_id"`
	Keywords        []string `avro:"keywords"`
	MetaDescription string   `avro:"meta_description"`
	ReleaseDate     string   `avro:"release_date"`
	Summary         string   `avro:"summary"`
	Title           string   `avro:"title"`
	Topics          []string `avro:"topics"`
	TraceID         string   `avro:"trace_id"`
}
