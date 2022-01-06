package clients

import (
	"context"

	datasetclient "github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/datasetapi.go -pkg mock . DatasetClient

// DatasetApiClient defines the zebedee client
type DatasetClient interface {
	Checker(context.Context, *healthcheck.CheckState) error
	// GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m datasetclient.Version, err error)
	GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version string) (m datasetclient.Metadata, err error)
}
