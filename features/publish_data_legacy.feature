@Legacy
Feature: Data extractor should listen to the relevant topic and publish extracted data for legacy (Zebedee) datasets

  Background:
    Given dp-dataset-api is healthy
    And zebedee is healthy

  Scenario: When searching for the extracted legacy data I get the expected result
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "some_type",
      "URI":  "some_uri",
      "description": {
        "title":     "title",
        "cdid":      "123",
        "datasetId": "456",
        "edition":   "something"
      }
    }
    """
    When the service starts
    And this "content-updated" avro event is queued, to be consumed
    """
    {
        "uri":           "some_uri",
        "data_type":      "legacy",
        "collection_id":  "123"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "uid":         "some_uri",
      "uri":         "some_uri",
      "title":       "title",
      "edition":     "something",
      "data_type":    "some_type",
      "search_index": "ons",
      "cdid":        "123",
      "dataset_id":   "456",
      "keywords":    [],
      "topics":      []
    }
    """

  Scenario: When no title is present, an item is not added to search
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "some_type",
      "URI":  "some_uri",
      "description": {
        "cdid":      "123",
        "datasetId": "456",
        "edition":   "something"
      }
    }
    """
    When the service starts
    And this "content-updated" avro event is queued, to be consumed
    """
    {
        "URI":           "some_uri",
        "DataType":      "legacy",
        "CollectionID":  "123"
    }
    """
    Then no search-data-import events are produced
    Then no search-content-deleted events are produced

  Scenario: When migrationLink exists and content type is NOT editorial, send search-content-deleted event
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "some_type",
      "uri":  "some_uri",
      "description": {
        "title":          "title",
        "cdid":           "123",
        "datasetId":      "456",
        "edition":        "something",
        "migrationLink":  "/migrated/content"
      }
    }
    """
    When the service starts
    And this "content-updated" avro event is queued, to be consumed
    """
    {
        "uri": "some_uri",
        "data_type": "legacy",
        "collection_id": "123"
    }
    """
    Then this search-content-deleted event is sent
    """
    {
        "uri": "some_uri",
        "collection_id":  "123",
        "search_index":  "ons"
    }
    """
    And no search-data-import events are produced

  Scenario: When migrationLink exists but content type is editorial, send search-data-import event
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "bulletin",
      "uri":  "some_uri",
      "description": {
        "title":          "title",
        "cdid":           "123",
        "datasetId":      "456",
        "edition":        "something",
        "migrationLink":  "/migrated/content"
      }
    }
    """
    When the service starts
    And this "content-updated" avro event is queued, to be consumed
    """
    {
        "uri":           "some_uri",
        "data_type":      "legacy",
        "collection_id":  "123"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "uid":         "some_uri",
      "uri":         "some_uri",
      "title":       "title",
      "edition":     "something",
      "data_type":    "bulletin",
      "search_index": "ons",
      "cdid":        "123",
      "dataset_id":   "456",
      "keywords":    [],
      "topics":      []
    }
    """
    And no search-content-deleted events are produced

  Scenario: When migrationLink does not exist, send search-data-import event
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "some_type",
      "URI":  "some_uri",
      "description": {
        "title":          "title",
        "cdid":           "123",
        "datasetId":      "456",
        "edition":        "something"
      }
    }
    """
    When the service starts
    And this "content-updated" avro event is queued, to be consumed
    """
    {
        "uri":           "some_uri",
        "data_type":      "legacy",
        "collection_id":  "123"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "uid":         "some_uri",
      "uri":         "some_uri",
      "title":       "title",
      "edition":     "something",
      "data_type":    "some_type",
      "search_index": "ons",
      "cdid":        "123",
      "dataset_id":   "456",
      "keywords":    [],
      "topics":      []
    }
    """
    And no search-content-deleted events are produced
