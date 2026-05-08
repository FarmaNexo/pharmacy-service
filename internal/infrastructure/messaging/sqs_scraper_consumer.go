// internal/infrastructure/messaging/sqs_scraper_consumer.go
package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/farmanexo/pharmacy-service/internal/application/handlers"
	"github.com/farmanexo/pharmacy-service/internal/domain/events"
	"github.com/farmanexo/pharmacy-service/pkg/config"
	"go.uber.org/zap"
)

// SQSScraperConsumer escucha la cola farmanexo-{env}-scraper-events,
// dedicada a PHARMACY_DISCOVERED + INVENTORY_DISCOVERED publicados por el
// scraper. La cola scraper-product-events (PRODUCT_DISCOVERED) la consume
// catalog-service.
type SQSScraperConsumer struct {
	sqsClient        *sqs.Client
	queueURL         string
	pharmacyHandler  *handlers.UpsertPharmacyFromEventHandler
	inventoryHandler *handlers.UpsertInventoryFromEventHandler
	logger           *zap.Logger
	stopCh           chan struct{}
}

func NewSQSScraperConsumer(
	awsCfg config.AWSConfig,
	sqsCfg config.SQSConfig,
	pharmacyHandler *handlers.UpsertPharmacyFromEventHandler,
	inventoryHandler *handlers.UpsertInventoryFromEventHandler,
	logger *zap.Logger,
) (*SQSScraperConsumer, error) {
	if sqsCfg.ScraperEventsQueueURL == "" {
		return nil, fmt.Errorf("SQS_SCRAPER_EVENTS_QUEUE_URL no está configurada")
	}

	optFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(awsCfg.Region),
	}
	if awsCfg.Endpoint != "" {
		optFns = append(optFns, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		))
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	sqsOptFns := []func(*sqs.Options){}
	if awsCfg.Endpoint != "" {
		sqsOptFns = append(sqsOptFns, func(o *sqs.Options) {
			o.BaseEndpoint = &awsCfg.Endpoint
		})
	}

	sqsClient := sqs.NewFromConfig(cfg, sqsOptFns...)

	logger.Info("SQS ScraperConsumer inicializado",
		zap.String("queue_url", sqsCfg.ScraperEventsQueueURL),
	)

	return &SQSScraperConsumer{
		sqsClient:        sqsClient,
		queueURL:         sqsCfg.ScraperEventsQueueURL,
		pharmacyHandler:  pharmacyHandler,
		inventoryHandler: inventoryHandler,
		logger:           logger,
		stopCh:           make(chan struct{}),
	}, nil
}

// Start arranca el polling en una goroutine.
func (c *SQSScraperConsumer) Start(ctx context.Context) {
	c.logger.Info("Iniciando SQS ScraperConsumer", zap.String("queue_url", c.queueURL))
	go c.pollMessages(ctx)
}

// Stop detiene el consumer.
func (c *SQSScraperConsumer) Stop() {
	close(c.stopCh)
	c.logger.Info("SQS ScraperConsumer detenido")
}

func (c *SQSScraperConsumer) pollMessages(ctx context.Context) {
	for {
		select {
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			c.receiveAndProcess(ctx)
		}
	}
}

func (c *SQSScraperConsumer) receiveAndProcess(ctx context.Context) {
	result, err := c.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &c.queueURL,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
	})
	if err != nil {
		c.logger.Warn("Error recibiendo mensajes SQS scraper-events", zap.Error(err))
		time.Sleep(5 * time.Second)
		return
	}
	for _, m := range result.Messages {
		c.processMessage(ctx, m)
	}
}

func (c *SQSScraperConsumer) processMessage(ctx context.Context, message sqstypes.Message) {
	if message.Body == nil {
		return
	}

	var envelope events.ScraperEvent
	if err := json.Unmarshal([]byte(*message.Body), &envelope); err != nil {
		c.logger.Error("Error deserializando ScraperEvent",
			zap.Error(err),
			zap.String("body", *message.Body),
		)
		c.deleteMessage(ctx, message)
		return
	}

	switch envelope.EventType {
	case events.ScraperEventPharmacyDiscovered:
		c.handlePharmacyDiscovered(ctx, envelope, message)
	case events.ScraperEventInventoryDiscovered:
		c.handleInventoryDiscovered(ctx, envelope, message)
	default:
		// La cola scraper-events solo debería traer PHARMACY/INVENTORY;
		// si aparece otro tipo es un bug de routing del publisher. Logueamos
		// y borramos para no atascar la cola.
		c.logger.Warn("Evento inesperado en scraper-events",
			zap.String("event_type", envelope.EventType),
			zap.String("source_id", envelope.SourceID),
		)
		c.deleteMessage(ctx, message)
	}
}

func (c *SQSScraperConsumer) handlePharmacyDiscovered(ctx context.Context, envelope events.ScraperEvent, message sqstypes.Message) {
	var data events.PharmacyDiscoveredData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		c.logger.Error("Error deserializando PharmacyDiscoveredData",
			zap.Error(err),
			zap.String("source_id", envelope.SourceID),
		)
		c.deleteMessage(ctx, message)
		return
	}

	if _, err := c.pharmacyHandler.Handle(ctx, data); err != nil {
		// Falla del UPSERT (ej. DB transient): NO borrar → SQS redrive.
		c.logger.Warn("UPSERT pharmacy falló, dejo el mensaje para redrive",
			zap.String("source_id", envelope.SourceID),
			zap.Error(err),
		)
		return
	}
	c.deleteMessage(ctx, message)
}

func (c *SQSScraperConsumer) handleInventoryDiscovered(ctx context.Context, envelope events.ScraperEvent, message sqstypes.Message) {
	var data events.InventoryDiscoveredData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		c.logger.Error("Error deserializando InventoryDiscoveredData",
			zap.Error(err),
			zap.String("source_id", envelope.SourceID),
		)
		c.deleteMessage(ctx, message)
		return
	}

	err := c.inventoryHandler.Handle(ctx, data)
	if err == nil {
		c.deleteMessage(ctx, message)
		return
	}
	// FailSoft = pharmacy o product aún no UPSERTeados; SQS redrive
	// para reintentar tras visibility timeout. Tras 3 reintentos sin
	// éxito, va a DLQ (esperado solo si hay un bug aguas arriba).
	if errors.Is(err, handlers.ErrFailSoft) {
		c.logger.Debug("INVENTORY redrive por dependencia faltante",
			zap.String("source_id", envelope.SourceID),
		)
		return
	}
	// Falla "real" (DB, etc): también dejamos para redrive.
	c.logger.Warn("UPSERT inventory falló, dejo el mensaje para redrive",
		zap.String("source_id", envelope.SourceID),
		zap.Error(err),
	)
}

func (c *SQSScraperConsumer) deleteMessage(ctx context.Context, message sqstypes.Message) {
	_, err := c.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &c.queueURL,
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		c.logger.Warn("Error borrando mensaje SQS", zap.Error(err))
	}
}
