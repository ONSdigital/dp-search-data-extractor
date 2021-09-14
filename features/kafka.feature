Feature: Kafka

  Scenario: Posting and checking a response
    When these kafka events are consumed:
            | URL           | DataType    | CollectionID |
            | "testURL.com" | "TestThing" | "Col123Test" |
    Then I should receive a kafka response