package event

// ContentPublished provides an avro structure for a Content Published event
type ContentPublished struct {
	URL string `avro:"url"`
	DataType string `avro:"data_type"`
	CollectionID string `avro:"collection_id"`
}
