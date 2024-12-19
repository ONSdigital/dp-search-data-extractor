@Dataset
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
    And this "content-updated" event is queued, to be consumed
    """
    {
        "URI":           "/datasets/cphi01/editions/timeseries/versions/version/metadata",
        "DataType":      "datasets",
        "CollectionID":  "123"
    }
    """
    Then this search-data-import event is sent
    """
      {
        "UID":         "cphi01-timeseries",
        "Edition":     "timeseries",
        "DataType":    "dataset_landing_page",
        "SearchIndex": "ons",
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
         { "id": "dim1", "label": "label 1 (11 categories)", "is_area_type": true },
         { "id": "dim3.0", "label": "label 3 (33 categories)" },
         { "id": "dim3.1", "label": "label 3 (40 categories)" }
      ]
    }
    """
    When the service starts
    And this "content-updated" event is queued, to be consumed
    """
    {
        "URI":          "/datasets/my-cantabular-dataset/editions/my-edition/versions/my-version/metadata",
        "DataType":     "datasets",
        "CollectionID": "123"
    }
    """
    Then this search-data-import event is sent
      """
      {
        "UID":         "my-cantabular-dataset-my-edition",
        "Edition":     "my-edition",
        "DataType":    "dataset_landing_page",
        "SearchIndex": "ons",
        "DatasetID":   "my-cantabular-dataset",
        "Keywords":    [ "keyword1", "keyword2" ],
        "ReleaseDate": "releasedate",
        "Summary":     "description",
        "Title":       "title",
        "Topics":      [],
        "PopulationType": {
          "Key":    "all-usual-residents-in-households",
          "AggKey": "all-usual-residents-in-households###All usual residents in households",
          "Name":   "UR_HH",
          "Label":  "All usual residents in households"
        },
        "Dimensions": [
          { "Key": "label-3", "AggKey": "label-3###label 3", "Name": "dim3.0,dim3.1", "Label": "label 3", "RawLabel": "label 3 (33 categories),label 3 (40 categories)"}
        ]
      }
      """
