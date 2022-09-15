package event

import (
	"context"
	"fmt"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out ./mock/producer.go -pkg mock . Marshaller

// Marshaller defines a type for marshalling the requested object into the required format.
type Marshaller interface {
	Marshal(s interface{}) ([]byte, error)
}

// SearchDataImportProducer produces kafka messages for instances which have been successfully processed.
type SearchDataImportProducer struct {
	Marshaller Marshaller
	Producer   kafka.IProducer
}

// SearchDataImport produce a kafka message for an instance which has been successfully processed.
func (p SearchDataImportProducer) SearchDataImport(ctx context.Context, event models.SearchDataImport) error {
	traceID := ctx.Value(kafka.TraceIDHeaderKey)
	bytes, err := p.Marshaller.Marshal(event)
	if err != nil {
		log.Fatal(ctx, "Marshaller.Marshal", err)
		return fmt.Errorf(fmt.Sprintf("Marshaller.Marshal returned an error: event=%v: %%w", event), err, log.Data{
			"request-id": traceID,
		})
	}

	p.Producer.Channels().Output <- bytes
	log.Info(ctx, "completed successfully", log.Data{
		"event":      event,
		"package":    "event.SearchDataImport",
		"request-id": traceID,
	})
	return nil
}
