package schema

import (
	"github.com/ONSdigital/dp-kafka/v2/avro"
)

var contentPublished = `{
  "type": "record",
  "name": "content-published",
  "fields": [
    {"name": "uri", "type": "string", "default": ""},
    {"name": "data_type", "type": "string", "default": ""},
    {"name": "collection_id", "type": "string", "default": ""}
  ]
}`

// ContentPublishedEvent is the Avro schema for Content Published messages.
var ContentPublishedEvent = &avro.Schema{
	Definition: contentPublished,
}

var searchDataImport = `{
  "type": "record",
  "name": "search-data-import",
  "fields": [
    {"name": "data_type", "type": "string", "default": ""},
    {"name": "job_id", "type": "string", "default": ""},
    {"name": "search_index", "type": "string", "default": ""},
    {"name": "cdid", "type": "string", "default": ""},
    {"name": "dataset_id", "type": "string", "default": ""},
    {"name": "keywords","type":["null",{"type":"array","items":"string"}]},
    {"name": "meta_description", "type": "string", "default": ""},
    {"name": "release_date", "type": "string", "default": ""},
    {"name": "summary", "type": "string", "default": ""},
    {"name": "title", "type": "string", "default": ""},
    {"name": "trace_id", "type": "string", "default": ""}
  ]
}`

// SearchDataImportEvent the Avro schema for search-data-import messages.
var SearchDataImportEvent = &avro.Schema{
	Definition: searchDataImport,
}

var searchDatasetVersionMetadataImport = `{
  "type": "record",
  "name": "search-data-import",
  "fields": [
    {"name": "data_id", "type": "string", "default": ""},
    {"name": "release_date", "type": "string", "default": ""},
    {"name": "latest_changes","type":["null",{"type":"array","items":"string"}]},
    {"name": "title", "type": "string", "default": ""},
    {"name": "description", "type": "string", "default": ""},
    {"name": "keywords","type":["null",{"type":"array","items":"string"}]},
    {"name": "release_frequency", "type": "string", "default": ""},
    {"name": "next_release", "type": "string", "default": ""},
    {"name": "unit_of_measure", "type": "string", "default": ""},
    {"name": "license", "type": "string", "default": ""},
    {"name": "national_statistic", "type": "string", "default": ""}
  ]
}`

// SearchDatasetVersionMetadataEvent the Avro schema for Search-Dataset-Version-Metadata messages.
var SearchDatasetVersionMetadataEvent = &avro.Schema{
	Definition: searchDatasetVersionMetadataImport,
}
