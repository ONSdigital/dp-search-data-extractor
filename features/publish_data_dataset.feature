Feature: Data extractor should listen to the relevant topic and publish extracted data for Dataset API datasets


  Scenario: When searching for the extracted dataset generic metadata I get the expected result
    Given dp-dataset-api is healthy
    And zebedee is healthy
    And the following metadata with dataset-id "cphi01", edition "timeseries" and version "version" is available in dp-dataset-api
    """
    {
        "release_date": "releasedate",
        "title":        "title",
        "description":  "description",
        "keywords":     [ "keyword1", "keyword2" ]
    }
    """

    When the service starts
    And this content-updated event is queued, to be consumed
      | URI                                                            | DataType | CollectionID |
      | /datasets/cphi01/editions/timeseries/versions/version/metadata | datasets | 123          |

    Then this search-data-import event is sent
    """
      {
        "UID":         "cphi01-timeseries",
        "Edition":     "timeseries",
        "DataType":    "dataset_landing_page",
        "SearchIndex": "ONS",
        "DatasetID":   "cphi01",
        "Keywords":    [ "keyword1", "keyword2" ],
        "ReleaseDate": "releasedate",
        "Summary":     "description",
        "Title":       "title",
        "Topics":      []
      }
      """


  Scenario: "When searching for the extracted dataset cantabular-type metadata I get the expected result"
    Given dp-dataset-api is healthy
    And zebedee is healthy
    And the following metadata with dataset-id "my-cantabular-dataset", edition "my-edition" and version "my-version" is available in dp-dataset-api
    """
    {
      "is_based_on": {
        "@type": "cantabular_flexible_table",
        "@id":   "UR_HH"
      },
      "release_date": "releasedate",
      "title":        "title",
      "description":  "description",
      "keywords":     [ "keyword1", "keyword2" ],
      "dimensions": [
         { "id": "dim1", "label": "label 1 (11 categories)" },
         { "id": "dim2", "label": "label 2 (22 categories)", "is_area_type": true },
         { "id": "dim3", "label": "label 3 (33 categories)" }
      ]
    }
    """

    When the service starts
    And this content-updated event is queued, to be consumed
      | URI                                                                              | DataType | CollectionID |
      | /datasets/my-cantabular-dataset/editions/my-edition/versions/my-version/metadata | datasets | 123          |

    Then this search-data-import event is sent
      """
      {
        "UID":         "my-cantabular-dataset-my-edition",
        "Edition":     "my-edition",
        "DataType":    "cantabular_flexible_table",
        "SearchIndex": "ONS",
        "DatasetID":   "my-cantabular-dataset",
        "Keywords":    [ "keyword1", "keyword2" ],
        "ReleaseDate": "releasedate",
        "Summary":     "description",
        "Title":       "title",
        "Topics":      [],
        "PopulationType": {
          "Name":  "UR_HH",
          "Label": "All usual residents in households"
        },
        "Dimensions": [
          { "Name": "dim1", "Label": "label 1", "RawLabel": "label 1 (11 categories)" },
          { "Name": "dim3", "Label": "label 3", "RawLabel": "label 3 (33 categories)" }
        ]
      }
      """
