package models

// ContentPublished provides an avro structure for a Content Published event
type ContentPublished struct {
	URI          string `avro:"uri" json:"uri"`
	DataType     string `avro:"data_type" json:"data_type"`
	CollectionID string `avro:"collection_id" json:"collection_id"`
	JobID        string `avro:"job_id" json:"job_id"`
	SearchIndex  string `avro:"search_index" json:"search_index"`
	TraceID      string `avro:"trace_id" json:"trace_id"`
}

// SearchDataImport provides event data for a search data import
type SearchDataImport struct {
	UID             string               `avro:"uid" json:"uid"`
	URI             string               `avro:"uri" json:"uri"`
	Edition         string               `avro:"edition" json:"edition"`
	DataType        string               `avro:"data_type" json:"data_type"`
	JobID           string               `avro:"job_id" json:"job_id"`
	SearchIndex     string               `avro:"search_index" json:"search_index"`
	CDID            string               `avro:"cdid" json:"cdid"`
	DatasetID       string               `avro:"dataset_id" json:"dataset_id"`
	Keywords        []string             `avro:"keywords" json:"keywords"`
	MetaDescription string               `avro:"meta_description" json:"meta_description"`
	ReleaseDate     string               `avro:"release_date" json:"release_date"`
	Summary         string               `avro:"summary" json:"summary"`
	Title           string               `avro:"title" json:"title"`
	Topics          []string             `avro:"topics" json:"topics"`
	TraceID         string               `avro:"trace_id" json:"trace_id"`
	DateChanges     []ReleaseDateDetails `avro:"date_changes" json:"date_changes"`
	Cancelled       bool                 `avro:"cancelled" json:"cancelled"`
	Finalised       bool                 `avro:"finalised" json:"finalised"`
	ProvisionalDate string               `avro:"provisional_date" json:"provisional_date"`
	CanonicalTopic  string               `avro:"canonical_topic" json:"canonical_topic"`
	Published       bool                 `avro:"published" json:"published"`
	Language        string               `avro:"language" json:"language"`
	Survey          string               `avro:"survey" json:"survey"`
	PopulationType  PopulationType       `avro:"population_type" json:"population_type"`
	Dimensions      []Dimension          `avro:"dimensions" json:"dimensions"`
}

// ReleaseDateDetails represents a date change of a release
type ReleaseDateDetails struct {
	ChangeNotice string `avro:"change_notice" json:"change_notice"`
	Date         string `avro:"previous_date" json:"previous_date"`
}

// Dimension represents the required information for each dataset dimension: name (unique ID) and label
// and an aggregation key which combines name and label
type Dimension struct {
	Key      string `avro:"key" json:"key"`
	AggKey   string `avro:"agg_key" json:"agg_key"`
	Name     string `avro:"name" json:"name"`
	Label    string `avro:"label" json:"label"`
	RawLabel string `avro:"raw_label" json:"raw_label"`
}

// PopulationType represents the population type name (unique ID) and label
// and an aggregation key which combines name and label
type PopulationType struct {
	Key    string `avro:"key" json:"key"`
	AggKey string `avro:"agg_key" json:"agg_key"`
	Name   string `avro:"name" json:"name"`
	Label  string `avro:"label" json:"label"`
}

// SearchContentUpdate - this represents a standard resource metadata model and json representation for API
type SearchContentUpdate struct {
	CanonicalTopic  string   `avro:"canonical_topic" json:"canonical_topic"`
	CDID            string   `avro:"cdid" json:"cdid"`
	ContentType     string   `avro:"content_type" json:"content_type"`
	DatasetID       string   `avro:"dataset_id" json:"dataset_id"`
	Edition         string   `avro:"edition" json:"edition"`
	Language        string   `avro:"language" json:"language"`
	MetaDescription string   `avro:"meta_description" json:"meta_description"`
	ReleaseDate     string   `avro:"release_date" json:"release_date"`
	Summary         string   `avro:"summary" json:"summary"`
	Survey          string   `avro:"survey" json:"survey"`
	Title           string   `avro:"title" json:"title"`
	Topics          []string `avro:"topics" json:"topics"`
	TraceID         string   `avro:"trace_id" json:"trace_id"`
	URI             string   `avro:"uri" json:"uri"`
	URIOld          string   `avro:"uri_old" json:"uri_old"`
	// These fields are only used for content_type=release
	Cancelled       bool                 `avro:"cancelled" json:"cancelled"`
	Finalised       bool                 `avro:"finalised" json:"finalised"`
	Published       bool                 `avro:"published" json:"published"`
	DateChanges     []ReleaseDateDetails `avro:"date_changes" json:"date_changes"`
	ProvisionalDate string               `avro:"provisional_date" json:"provisional_date"`
}

// SearchContentDeleted represents event data for search-content-deleted
type SearchContentDeleted struct {
	URI          string `avro:"uri" json:"uri"`
	CollectionID string `avro:"collection_id" json:"collection_id"`
	SearchIndex  string `avro:"search_index" json:"search_index"`
	TraceID      string `avro:"trace_id" json:"trace_id"`
}
