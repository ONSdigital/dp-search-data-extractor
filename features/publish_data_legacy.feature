@Legacy
Feature: Data extractor should listen to the relevant topic and publish extracted data for legacy (Zebedee) datasets

  Background:
    Given dp-dataset-api is healthy
    And zebedee is healthy

  Scenario: When searching for the extracted legacy data I get the expected result
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "legacy",
      "URI":  "some_uri",
      "description": {
        "title":     "title",
        "data_type": "legacy",
        "cdid":      "123",
        "datasetId": "456",
        "edition":   "something"
      }
    }
    """
    When the service starts
    And this "content-updated" event is queued, to be consumed
    """
    {
        "URI":           "some_uri",
        "DataType":      "legacy",
        "CollectionID":  "123"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "UID":         "some_uri",
      "URI":         "some_uri",
      "Title":       "title",
      "Edition":     "something",
      "DataType":    "legacy",
      "SearchIndex": "ons",
      "CDID":        "123",
      "DatasetID":   "456",
      "Keywords":    [],
      "Topics":      []
    }
    """

  Scenario: When no title is present, an item is not added to search
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "legacy",
      "URI":  "some_uri",
      "description": {
        "data_type": "legacy",
        "cdid":      "123",
        "datasetId": "456",
        "edition":   "something"
      }
    }
    """
    When the service starts
    And this "content-updated" event is queued, to be consumed
    """
    {
        "URI":           "some_uri",
        "DataType":      "legacy",
        "CollectionID":  "123"
    }
    """
    Then no search-data-import events are produced

  Scenario: When migrationLink exists and content type is NOT editorial, send search-content-deleted event
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "legacy",
      "URI":  "some_uri",
      "description": {
        "title":          "title",
        "data_type":      "legacy",
        "cdid":           "123",
        "datasetId":      "456",
        "edition":        "something",
        "migrationLink":  "/migrated/content"
      }
    }
    """
    When the service starts
    And this "content-updated" event is queued, to be consumed
    """
    {
        "URI":           "some_uri",
        "DataType":      "legacy",
        "CollectionID":  "123"
    }
    """
    Then this search-content-deleted event is sent
    """
    {
      "URI": "some_uri",
      "CollectionID":  "123",
      "SearchIndex":  "ons"
    }
    """
    And no search-data-import events are produced

  Scenario: When migrationLink exists but content type is editorial, send search-data-import event
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "bulletin",
      "URI":  "some_uri",
      "description": {
        "title":          "title",
        "data_type":      "legacy",
        "cdid":           "123",
        "datasetId":      "456",
        "edition":        "something",
        "migrationLink":  "/migrated/content"
      }
    }
    """
    When the service starts
    And this "content-updated" event is queued, to be consumed
    """
    {
        "URI":           "some_uri",
        "DataType":      "legacy",
        "CollectionID":  "123"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "UID":         "some_uri",
      "URI":         "some_uri",
      "Title":       "title",
      "Edition":     "something",
      "DataType":    "bulletin",
      "SearchIndex": "ons",
      "CDID":        "123",
      "DatasetID":   "456",
      "Keywords":    [],
      "Topics":      []
    }
    """
    And no search-content-deleted events are produced

  Scenario: When migrationLink does not exist, send search-data-import event
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "legacy",
      "URI":  "some_uri",
      "description": {
        "title":          "title",
        "data_type":      "legacy",
        "cdid":           "123",
        "datasetId":      "456",
        "edition":        "something"
      }
    }
    """
    When the service starts
    And this "content-updated" event is queued, to be consumed
    """
    {
        "URI":           "some_uri",
        "DataType":      "legacy",
        "CollectionID":  "123"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "UID":         "some_uri",
      "URI":         "some_uri",
      "Title":       "title",
      "Edition":     "something",
      "DataType":    "legacy",
      "SearchIndex": "ons",
      "CDID":        "123",
      "DatasetID":   "456",
      "Keywords":    [],
      "Topics":      []
    }
    """
    And no search-content-deleted events are produced
