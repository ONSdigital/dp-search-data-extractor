package redirects

import (
	"context"
	"errors"
	"testing"

	redirectModels "github.com/ONSdigital/dis-redirect-api/models"
	redirectAPI "github.com/ONSdigital/dis-redirect-api/sdk/go"
	apiError "github.com/ONSdigital/dis-redirect-api/sdk/go/errors"
	redirectClientMock "github.com/ONSdigital/dis-redirect-api/sdk/go/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetPreviousURIs(t *testing.T) {
	Convey("Given a redirect client with dataset redirects", t, func() {
		client := &redirectClientMock.ClienterMock{GetRedirectsFunc: func(_ context.Context, _ redirectAPI.Options) (*redirectModels.Redirects, apiError.Error) {
			return &redirectModels.Redirects{
				RedirectList: []redirectModels.Redirect{{From: "/economy/nationalaccounts/datasets/teststaticdataset"}, {From: "/economy/environmentalaccounts/datasets/teststaticdataset"}, {From: "/economy/nationalaccounts/datasets/teststaticdataset"}, {From: ""}},
				NextCursor:   "0",
			}, nil
		}}

		Convey("When previous URIs are requested for the dataset", func() {
			previousURIs, err := GetPreviousURIs(context.Background(), client, "/datasets/teststaticdataset", "testToken")

			Convey("Then the request targets the dataset and returns sorted unique paths", func() {
				So(err, ShouldBeNil)
				So(client.GetRedirectsCalls(), ShouldHaveLength, 1)
				So(previousURIs, ShouldResemble, []string{"/economy/environmentalaccounts/datasets/teststaticdataset", "/economy/nationalaccounts/datasets/teststaticdataset"})
			})
		})

		Convey("When the redirect client returns no redirects", func() {
			client.GetRedirectsFunc = func(_ context.Context, _ redirectAPI.Options) (*redirectModels.Redirects, apiError.Error) {
				return &redirectModels.Redirects{NextCursor: "0"}, nil
			}
			previousURIs, err := GetPreviousURIs(context.Background(), client, "/datasets/teststaticdataset", "testToken")

			Convey("Then the result is empty", func() {
				So(err, ShouldBeNil)
				So(client.GetRedirectsCalls(), ShouldHaveLength, 1)
				So(previousURIs, ShouldResemble, []string{})
			})
		})

		Convey("When the redirect lookup fails", func() {
			client.GetRedirectsFunc = func(_ context.Context, _ redirectAPI.Options) (*redirectModels.Redirects, apiError.Error) {
				return nil, apiError.StatusError{Err: errors.New("redirect API unavailable")}
			}
			previousURIs, err := GetPreviousURIs(context.Background(), client, "/datasets/teststaticdataset", "")

			Convey("Then the lookup error includes the target path", func() {
				So(previousURIs, ShouldBeNil)
				So(client.GetRedirectsCalls(), ShouldHaveLength, 1)
				So(err.Error(), ShouldContainSubstring, "get redirects for target URI /datasets/teststaticdataset: redirect API unavailable")
			})
		})

		Convey("When there is no redirect client", func() {
			previousURIs, err := GetPreviousURIs(context.Background(), nil, "/datasets/teststaticdataset", "")

			Convey("Then the missing client error is returned without a lookup", func() {
				So(previousURIs, ShouldBeNil)
				So(client.GetRedirectsCalls(), ShouldHaveLength, 0)
				So(err, ShouldNotBeNil)
			})
		})
	})
}
