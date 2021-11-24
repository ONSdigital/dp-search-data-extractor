package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// ContentPublishedHandler struct to hold handle for zebedee client and the producer
type ContentPublishedHandler struct {
	ZebedeeCli clients.ZebedeeClient
	Producer   event.SearchDataImportProducer
}

// Handle takes a single event.
func (h *ContentPublishedHandler) Handle(ctx context.Context, event *models.ContentPublished, keywordsLimit string) (err error) {

	traceID := request.NewRequestID(16)

	logData := log.Data{
		"event": event,
	}
	log.Info(ctx, "event handler called", logData)

	contentPublished, err := h.ZebedeeCli.GetPublishedData(ctx, event.URI)
	if err != nil {
		return err
	}

	logData = log.Data{
		"contentPublished": string(contentPublished),
	}
	log.Info(ctx, "zebedee response ", logData)

	//byte slice to Json & unMarshall Json
	var zebedeeData models.ZebedeeData
	err = json.Unmarshal(contentPublished, &zebedeeData)
	if err != nil {
		log.Fatal(ctx, "error while attempting to unmarshal zebedee response into zebedeeData", err)
		return err
	}

	//keywords validation
	logData = log.Data{
		"keywords":      zebedeeData.Description.Keywords,
		"keywordsLimit": keywordsLimit,
	}
	validKeywords, err := ValidateKeywords(zebedeeData.Description.Keywords, keywordsLimit)
	if err != nil {
		log.Info(ctx, "failed to get keywords limit", logData)
		return err
	}

	//Mapping Json to Avro
	searchData := models.SearchDataImport{
		DataType:        zebedeeData.DataType,
		JobID:           "",
		SearchIndex:     "ONS",
		CDID:            zebedeeData.Description.CDID,
		DatasetID:       zebedeeData.Description.DatasetID,
		Keywords:        validKeywords,
		MetaDescription: zebedeeData.Description.MetaDescription,
		Summary:         zebedeeData.Description.Summary,
		ReleaseDate:     zebedeeData.Description.ReleaseDate,
		Title:           zebedeeData.Description.Title,
		TraceID:         traceID,
	}

	//Marshall Avro and sending message
	if err := h.Producer.SearchDataImport(ctx, searchData); err != nil {
		log.Fatal(ctx, "error while attempting to send SearchDataImport event to producer", err)
		return err
	}

	log.Info(ctx, "event successfully handled", logData)
	return nil
}

// incoming keywords validation
func ValidateKeywords(keywords []string, keywordsLimit string) ([]string, error) {

	var strArray []string
	validKeywords := make([]string, 0)

	for i := range keywords {
		strArray = strings.Split(keywords[i], ",")

		for j := range strArray {
			keyword := strings.TrimSpace(strArray[j])
			validKeywords = append(validKeywords, keyword)
		}
	}

	intKeywordsLimit, err := strconv.Atoi(keywordsLimit) //ASCII to integer
	if err != nil {
		return keywords, err
	}

	if intKeywordsLimit == 0 {
		return []string{""}, nil
	}

	if (len(validKeywords) < intKeywordsLimit) || (intKeywordsLimit == -1) {
		return validKeywords, nil
	}

	return validKeywords[:intKeywordsLimit], nil
}
