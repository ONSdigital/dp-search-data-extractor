Feature: Data extractor should listen to the relevant topic and publish extracted data for legacy (Zebedee) datasets


  Scenario: When searching for the extracted legacy data I get the expected result
    Given dp-dataset-api is healthy
    And zebedee is healthy
    And the following published data for uri "some_uri" is available in zebedee
    """
    {
      "type": "legacy",
      "URI": "some_uri",
      "description": {
        "data_type": "legacy",
        "cdid":      "123",
        "datasetId": "456",
        "edition":   "something"
      }
    }
    """

    When the service starts
    And this content-updated event is queued, to be consumed
      | URI                | DataType             |   CollectionID  |
      | some_uri           | legacy               |    123          |

    Then this search-data-import event is sent
    """
    {
      "UID":         "some_uri",
      "URI":         "some_uri",
      "Edition":     "something",
      "DataType":    "legacy",
      "SearchIndex": "ons",
      "CDID":        "123",
      "DatasetID":   "456",
      "Keywords":    [],
      "Topics":      []
    }
    """
