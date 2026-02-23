// internal/infrastructure/messaging/sqs_publisher.go
package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/farmanexo/pharmacy-service/internal/domain/events"
	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	"github.com/farmanexo/pharmacy-service/pkg/config"
	"go.uber.org/zap"
)

type SQSEventPublisher struct {
	sqsClient *sqs.Client
	queueURL  string
	logger    *zap.Logger
}

func NewSQSEventPublisher(awsCfg config.AWSConfig, sqsCfg config.SQSConfig, logger *zap.Logger) (*SQSEventPublisher, error) {
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
	logger.Info("SQS EventPublisher inicializado",
		zap.String("region", awsCfg.Region),
		zap.String("queue_url", sqsCfg.PharmacyEventsQueueURL),
	)

	return &SQSEventPublisher{sqsClient: sqsClient, queueURL: sqsCfg.PharmacyEventsQueueURL, logger: logger}, nil
}

func (p *SQSEventPublisher) Publish(ctx context.Context, event events.PharmacyEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}
	msgBody := string(body)
	input := &sqs.SendMessageInput{QueueUrl: &p.queueURL, MessageBody: &msgBody}
	_, err = p.sqsClient.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("error enviando mensaje a SQS: %w", err)
	}
	p.logger.Debug("Evento publicado en SQS", zap.String("event_type", event.EventType), zap.String("pharmacy_id", event.PharmacyID))
	return nil
}

var _ services.EventPublisher = (*SQSEventPublisher)(nil)
