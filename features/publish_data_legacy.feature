Feature: Data extractor should listen to the relevant topic and publish extracted data for legacy (Zebedee) datasets

  Scenario: When searching for the extracted legacy data I get the expected result
    Given I send a kafka event to content published topic
      | URI                | DataType             |   CollectionID  |
      | some_uri           | legacy               |    123          |
    When The kafka event is processed
    Then I should receive a kafka event to search-data-import topic with the following fields
      | UID        | Edition   | DataType | SearchIndex | CDID | DatasetID | Keywords | Topics |
      | something  | something | legacy   | ONS         | 123  | 456       |          |        |
