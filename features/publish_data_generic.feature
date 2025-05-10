@Generic
Feature:Data extractor should listen to the relevant topic and publish extracted data for Generic Upstream Services

  Scenario: When a valid `search-content-updated` of type release event is consumed, a `search-data-import` event is published
    Given dp-dataset-api is healthy
    And zebedee is healthy
    When the service starts
    And this "search-content-updated" event is queued, to be consumed
    """
    {
      "URI":           "/some/uri",
      "Title":         "title",
      "Edition":       "something",
      "ContentType":   "release",
      "CDID":          "123",
      "DatasetID":     "456"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "UID":         "/some/uri",
      "URI":         "/some/uri",
      "Title":       "title",
      "Edition":     "something",
      "DataType":    "release",
      "CDID":        "123",
      "DatasetID":   "456",
      "Keywords":    [],
      "Topics":      []
    }
    """


  Scenario: When a valid `search-content-updated` that is not a release event is consumed, a `search-data-import` event is published
    Given dp-dataset-api is healthy
    And zebedee is healthy
    When the service starts
    And this "search-content-updated" event is queued, to be consumed
    """
    {
      "URI":           "/some/uri",
      "Title":         "title",
      "Edition":       "something",
      "ContentType":   "article",
      "CDID":          "123",
      "DatasetID":     "456"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "UID":         "/some/uri",
      "URI":         "/some/uri",
      "Title":       "title",
      "Edition":     "something",
      "DataType":    "article",
      "CDID":        "123",
      "DatasetID":   "456",
      "Keywords":    [],
      "Topics":      []
    }
    """
