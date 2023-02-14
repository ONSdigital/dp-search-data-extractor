package schema

import (
	"github.com/ONSdigital/dp-kafka/v2/avro"
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

var searchDataImport = `

	{
		"type": "record",
		"name": "search-data-import",
		"fields": [{
				"name": "uid",
				"type": "string",
				"default": ""
			},
			{
				"name": "uri",
				"type": "string",
				"default": ""
			},
			{
				"name": "data_type",
				"type": "string",
				"default": ""
			},
			{
				"name": "job_id",
				"type": "string",
				"default": ""
			},
			{
				"name": "search_index",
				"type": "string",
				"default": ""
			},
			{
				"name": "cdid",
				"type": "string",
				"default": ""
			},
			{
				"name": "dataset_id",
				"type": "string",
				"default": ""
			},
			{
				"name": "edition",
				"type": "string",
				"default": ""
			},
			{
				"name": "keywords",
				"type": {
					"type": "array",
					"items": "string"
				}
			},
			{
				"name": "meta_description",
				"type": "string",
				"default": ""
			},
			{
				"name": "release_date",
				"type": "string",
				"default": ""
			},
			{
				"name": "summary",
				"type": "string",
				"default": ""
			},
			{
				"name": "title",
				"type": "string",
				"default": ""
			},
			{
				"name": "topics",
				"type": {
					"type": "array",
					"items": "string"
				}
			},
			{
				"name": "trace_id",
				"type": "string",
				"default": ""
			},
			{
				"name": "cancelled",
				"type": "boolean",
				"default": false
			},
			{
				"name": "finalised",
				"type": "boolean",
				"default": false
			},
			{
				"name": "published",
				"type": "boolean",
				"default": false
			},
			{
				"name": "language",
				"type": "string",
				"default": ""
			},
			{
				"name": "survey",
				"type": "string",
				"default": ""
			},
			{
				"name": "canonical_topic",
				"type": "string",
				"default": ""
			},
			{
				"name": "date_changes",
				"type": {
					"type": "array",
					"items": {
						"name": "ReleaseDateDetails",
						"type": "record",
            "default": "null",
						"fields": [{
								"name": "change_notice",
								"type": "string",
								"default": ""
							},
							{
								"name": "previous_date",
								"type": "string",
								"default": ""
							}
						]
					}
				}
			},
			{
				"name": "provisional_date",
				"type": "string",
				"default": ""
			},
			{
				"name": "dimensions",
	      "default": null,
				"type": {
					"type": "array",
					"items": {
						"name": "VersionDimension",
						"type": "record",
						"fields": [{
								"name": "id",
								"type": "string",
								"default": ""
							},
							{
								"name": "name",
								"type": "string",
								"default": ""
							},
							{
								"name": "description",
								"type": "string",
								"default": ""
							},
							{
								"name": "label",
								"type": "string",
								"default": ""
							},
							{
								"name": "href",
								"type": "string",
								"default": ""
							},
							{
								"name": "variable",
								"type": "string",
								"default": ""
							},
							{
								"name": "number_of_options",
								"type": "int"
							}
						]
					}
				}
	    }
		]
	}`

/*
var searchDataImport = `
[{
  "type": "record",
  "name": "Link",
  "fields": [{
      "name": "id",
      "type": "string",
      "default": ""
    },
    {
      "name": "href",
      "type": "string",
      "default": ""
    }
  ]
},
{
  "name": "Links",
  "type": "record",
  "fields": [{
      "name": "access_rights",
      "type": "Link"
    },
    {
      "name": "dataset",
      "type": "Link"
    },
    {
      "name": "dimensions",
      "type": "Link"
    },
    {
      "name": "edition",
      "type": "Link"
    },
    {
      "name": "editions",
      "type": "Link"
    },
    {
      "name": "latest_version",
      "type": "Link"
    },
    {
      "name": "versions",
      "type": "Link"
    },
    {
      "name": "self",
      "type": "Link"
    },
    {
      "name": "code_list",
      "type": "Link"
    },
    {
      "name": "options",
      "type": "Link"
    },
    {
      "name": "version",
      "type": "Link"
    },
    {
      "name": "taxonomy",
      "type": "Link"
    },
    {
      "name": "job",
      "type": "Link"
    },
    {
      "name": "code",
      "type": "Link"
    },
    {
      "name": "code",
      "type": "Link"
    }
  ]

},
{
  "type": "record",
  "name": "search-data-import",
  "fields": [{
      "name": "uid",
      "type": "string",
      "default": ""
    },
    {
      "name": "uri",
      "type": "string",
      "default": ""
    },
    {
      "name": "data_type",
      "type": "string",
      "default": ""
    },
    {
      "name": "job_id",
      "type": "string",
      "default": ""
    },
    {
      "name": "search_index",
      "type": "string",
      "default": ""
    },
    {
      "name": "cdid",
      "type": "string",
      "default": ""
    },
    {
      "name": "dataset_id",
      "type": "string",
      "default": ""
    },
    {
      "name": "edition",
      "type": "string",
      "default": ""
    },
    {
      "name": "keywords",
      "type": {
        "type": "array",
        "items": "string"
      }
    },
    {
      "name": "meta_description",
      "type": "string",
      "default": ""
    },
    {
      "name": "release_date",
      "type": "string",
      "default": ""
    },
    {
      "name": "summary",
      "type": "string",
      "default": ""
    },
    {
      "name": "title",
      "type": "string",
      "default": ""
    },
    {
      "name": "topics",
      "type": {
        "type": "array",
        "items": "string"
      }
    },
    {
      "name": "trace_id",
      "type": "string",
      "default": ""
    },
    {
      "name": "cancelled",
      "type": "boolean",
      "default": false
    },
    {
      "name": "finalised",
      "type": "boolean",
      "default": false
    },
    {
      "name": "published",
      "type": "boolean",
      "default": false
    },
    {
      "name": "language",
      "type": "string",
      "default": ""
    },
    {
      "name": "survey",
      "type": "string",
      "default": ""
    },
    {
      "name": "canonical_topic",
      "type": "string",
      "default": ""
    },
    {
      "name": "date_changes",
      "type": {
        "type": "array",
        "items": {
          "name": "ReleaseDateDetails",
          "type": "record",
          "fields": [{
              "name": "change_notice",
              "type": "string",
              "default": ""
            },
            {
              "name": "previous_date",
              "type": "string",
              "default": ""
            }
          ]
        }
      }
    },
    {
      "name": "provisional_date",
      "type": "string",
      "default": ""
    },
    {
      "name": "dimensions",
      "default": null,
      "type": {
        "type": "array",
        "items": {
          "name": "VersionDimension",
          "type": "record",
          "fields": [{
              "name": "id",
              "type": "string",
              "default": ""
            },
            {
              "name": "name",
              "type": "string",
              "default": ""
            },
            {
              "name": "links",
              "default": null,
              "type": "Links"
            },
            {
              "name": "description",
              "type": "string",
              "default": ""
            },
            {
              "name": "label",
              "type": "string",
              "default": ""
            },
            {
              "name": "href",
              "type": "string",
              "default": ""
            },
            {
              "name": "variable",
              "type": "string",
              "default": ""
            },
            {
              "name": "number_of_options",
              "type": "int",
              "default": 0
            }
          ]
        }
      }
    }
  ]
}
]
`
*/
// SearchDataImportEvent the Avro schema for search-data-import messages.
var SearchDataImportEvent = &avro.Schema{
	Definition: searchDataImport,
}
