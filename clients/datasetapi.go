package clients

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/datasetapi.go -pkg mock . DatasetClient

// DatasetApiClient defines the zebedee client
type DatasetClient interface {
	Checker(context.Context, *healthcheck.CheckState) error
	GetEdition(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition string) (m dataset.Edition, err error)
}
