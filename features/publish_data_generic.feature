Feature:Data extractor should listen to the relevant topic and publish extracted data for Generic Upstream Services

  @Generic
  Scenario: When a valid `search-content-updated` event is consumed, a `search-data-import` event is published
    Given dp-dataset-api is healthy
    And zebedee is healthy

    When the service starts
    And this "search-content-updated" event is queued, to be consumed
    """
    {
      "URI":           "/some/uri",
      "DataType":      "release",
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
      "DataType":    "release",
      "SearchIndex": "ons",
      "CDID":        "123",
      "DatasetID":   "456",
      "Keywords":    [],
      "Topics":      []
    }
    """