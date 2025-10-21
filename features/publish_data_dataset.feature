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
    And this "content-updated" avro event is queued, to be consumed
    """
    {
        "uri":           "/datasets/cphi01/editions/timeseries/versions/version/metadata",
        "data_type":      "datasets",
        "collection_id":  "123"
    }
    """
    Then this search-data-import event is sent
    """
      {
        "uid":         "cphi01-timeseries",
        "edition":     "timeseries",
        "data_type":    "dataset_landing_page",
        "search_index": "ons",
        "dataset_id":   "cphi01",
        "keywords":    [ "keyword1", "keyword2" ],
        "summary":     "description",
        "title":       "title",
        "release_date": "releasedate",
        "topics":      []
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
    And this "content-updated" avro event is queued, to be consumed
    """
    {
        "uri":          "/datasets/my-cantabular-dataset/editions/my-edition/versions/my-version/metadata",
        "data_type":     "datasets",
        "collection_id": "123"
    }
    """
    Then this search-data-import event is sent
      """
      {
        "uid":         "my-cantabular-dataset-my-edition",
        "edition":     "my-edition",
        "data_type":    "dataset_landing_page",
        "search_index": "ons",
        "dataset_id":   "my-cantabular-dataset",
        "keywords":    [ "keyword1", "keyword2" ],
        "release_date": "releasedate",
        "summary":     "description",
        "title":       "title",
        "topics":      [],
        "population_type": {
          "key":    "all-usual-residents-in-households",
          "agg_key": "all-usual-residents-in-households###All usual residents in households",
          "name":   "UR_HH",
          "label":  "All usual residents in households"
        },
        "dimensions": [
          { "key": "label-3", "agg_key": "label-3###label 3", "name": "dim3.0,dim3.1", "label": "label 3", "raw_label": "label 3 (33 categories),label 3 (40 categories)"}
        ]
      }
      """
