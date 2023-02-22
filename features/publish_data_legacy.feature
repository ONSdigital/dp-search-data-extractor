# Feature: Data extractor should listen to the relevant topic and publish extracted data for legacy (Zebedee) datasets


#   Scenario: When searching for the extracted legacy data I get the expected result
#     Given dp-dataset-api is healthy
#     And zebedee is healthy
#     And the following published data for uri "some_uri" is available in zebedee
#     """
#     {
#         "data_type": "legacy",
#         "cdid": "123",
#         "dataset_id": "456",
#         "edition": "something"
#     }
#     """

#     When the service starts
#     And this content-updated event is queued, to be consumed
#       | URI                | DataType             |   CollectionID  |
#       | some_uri           | legacy               |    123          |

#     Then I should receive a kafka event to search-data-import topic with the following fields
#       | UID        | Edition   | DataType | SearchIndex | CDID | DatasetID | Keywords | Topics |
#       | something  | something | legacy   | ONS         | 123  | 456       |          |        |
