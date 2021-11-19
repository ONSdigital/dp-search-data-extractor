package handler

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/log.go/v2/log"
)

const keywordsSize = 5

// ContentPublishedHandler struct to hold handle for zebedee client and the producer
type ContentPublishedHandler struct {
	ZebedeeCli clients.ZebedeeClient
	Producer   event.SearchDataImportProducer
}

// Handle takes a single event.
func (h *ContentPublishedHandler) Handle(ctx context.Context, event *models.ContentPublished) (err error) {

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
	validKeywords := ValidateKeywords(zebedeeData.Description.Keywords)

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
func ValidateKeywords(keywords []string) []string {

	var strArray []string
	validKeywords := make([]string, 0)

	for i := range keywords {
		strArray = strings.Split(keywords[i], ",")

		for j := range strArray {
			keyword := strings.TrimSpace(strArray[j])
			validKeywords = append(validKeywords, keyword)
		}
	}

	if len(validKeywords) < keywordsSize {
		return validKeywords
	}

	return validKeywords[:keywordsSize]
}
