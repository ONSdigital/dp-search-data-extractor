package schema

import (
	"github.com/ONSdigital/dp-kafka/v3/avro"
)

var contentUpdated = `{
  "type": "record",
  "name": "content-updated",
  "fields": [
    {"name": "uri", "type": "string", "default": ""},
    {"name": "data_type", "type": "string", "default": ""},
    {"name": "collection_id", "type": "string", "default": ""},
    {"name": "job_id", "type": "string", "default": ""},
    {"name": "search_index", "type": "string", "default": ""},
    {"name": "trace_id", "type": "string", "default": ""}
  ]
}`

// ContentPublishedEvent is the Avro schema for Content Published messages.
var ContentPublishedEvent = &avro.Schema{
	Definition: contentUpdated,
}

var searchDataImport = `{
  "type": "record",
  "name": "search-data-import",
  "fields": [
    {"name": "uid", "type": "string", "default": ""},
    {"name": "uri", "type": "string", "default": ""},
    {"name": "data_type", "type": "string", "default": ""},
    {"name": "job_id", "type": "string", "default": ""},
    {"name": "search_index", "type": "string", "default": ""},
    {"name": "cdid", "type": "string", "default": ""},
    {"name": "dataset_id", "type": "string", "default": ""},
    {"name": "edition", "type": "string", "default": ""},
    {"name": "keywords", "type": {"type":"array","items":"string"}, "default": []},
    {"name": "meta_description", "type": "string", "default": ""},
    {"name": "release_date", "type": "string", "default": ""},
    {"name": "summary", "type": "string", "default": ""},
    {"name": "title", "type": "string", "default": ""},
    {"name": "topics", "type": {"type":"array","items":"string"}, "default": []},
    {"name": "trace_id", "type": "string", "default": ""},
    {"name": "cancelled", "type": "boolean", "default": false},
    {"name": "finalised", "type": "boolean", "default": false},
    {"name": "published", "type": "boolean", "default": false},
    {"name": "language", "type": "string", "default": ""},
    {"name": "survey",   "type": "string", "default": ""},
    {"name": "canonical_topic",   "type": "string", "default": ""},
    {"name": "date_changes", "type": {"type":"array","items":{
      "name": "ReleaseDateDetails",
      "type" : "record",
      "fields" : [
        {"name": "change_notice", "type": "string", "default": ""},
        {"name": "previous_date", "type": "string", "default": ""}
      ]
    }}},
    {"name": "provisional_date", "type": "string", "default": ""},
    {"name": "dimensions", "type": {"type": "array", "items": {
      "name": "Dimension",
      "type" : "record",
      "fields": [
        { "name": "key", "type": "string", "default": "" },
        { "name": "agg_key", "type": "string", "default": "" },
        { "name": "name", "type": "string", "default": "" },
        { "name": "label", "type": "string", "default": "" },
        { "name": "raw_label", "type": "string", "default": "" }
      ]
    }}, "default": []},
    {"name": "population_type", "type": {
      "name": "PopulationType", "type": "record", "fields": [
        { "name": "key", "type": "string", "default": "" },
        { "name": "agg_key", "type": "string", "default": "" },
        { "name": "name", "type": "string", "default": ""},
        { "name": "label", "type": "string", "default": ""}
      ]
    }, "default": {"key": "", "agg_key": "", "name": "", "label": ""}}
  ]
}`

// SearchDataImportEvent the Avro schema for search-data-import messages.
var SearchDataImportEvent = &avro.Schema{
	Definition: searchDataImport,
}

var searchContentUpdate = `{
  "type": "record",
  "name": "search-content-updated",
  "fields": [
    {"name": "canonical_topic", "type": "string", "default": ""},
    {"name": "cdid", "type": "string", "default": ""},
    {"name": "content_type", "type": "string", "default": ""},
    {"name": "dataset_id", "type": "string", "default": ""},
    {"name": "edition", "type": "string", "default": ""},
    {"name": "language", "type": "string", "default": ""},
    {"name": "meta_description", "type": "string", "default": ""},
    {"name": "release_date", "type": "string", "default": ""},
    {"name": "search_index", "type": "string", "default": ""},
    {"name": "summary", "type": "string", "default": ""},
    {"name": "survey", "type": "string", "default": ""},
    {"name": "title", "type": "string", "default": ""},
    {"name": "topics", "type": {"type": "array", "items": "string"}, "default": []},
    {"name": "trace_id", "type": "string", "default": ""},
    {"name": "uri", "type": "string", "default": ""},
    {"name": "uri_old", "type": "string", "default": ""},
    {"name": "cancelled", "type": "boolean", "default": false},
    {"name": "finalised", "type": "boolean", "default": false},
    {"name": "published", "type": "boolean", "default": false},
    {"name": "date_changes", "type": { "type": "array", "items": {
      "name": "ReleaseDateDetails",
      "type": "record",
      "fields": [
        { "name":"change_notice", "type":"string" },
        { "name":"previous_date", "type":"string" }
      ] 
    }}},
    {"name": "provisional_date", "type": "string", "default": ""}
  ]
}`

// SearchContentUpdateEvent the Avro schema for search-content-updated messages.
var SearchContentUpdateEvent = &avro.Schema{
	Definition: searchContentUpdate,
}

var searchContentDelete = `{
  "type": "record",
  "name": "search-content-deleted",
  "fields": [
    {"name": "uri", "type": "string", "default": ""},
    {"name": "collection_id", "type": "string", "default": ""},
    {"name": "search_index", "type": "string", "default": ""},
    {"name": "trace_id", "type": "string", "default": ""}
 ]
}`

var SearchContentDeletedEvent = &avro.Schema{
	Definition: searchContentDelete,
}
