// internal/infrastructure/storage/s3_storage.go
package storage

import (
	"context"
	"fmt"
	"io"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	pkgconfig "github.com/farmanexo/pharmacy-service/pkg/config"
	"go.uber.org/zap"
)

type S3FileStorage struct {
	s3Client *s3.Client
	endpoint string
	region   string
	logger   *zap.Logger
}

func NewS3FileStorage(awsCfg pkgconfig.AWSConfig, logger *zap.Logger) (*S3FileStorage, error) {
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

	s3OptFns := []func(*s3.Options){}
	if awsCfg.Endpoint != "" {
		s3OptFns = append(s3OptFns, func(o *s3.Options) {
			o.BaseEndpoint = &awsCfg.Endpoint
			o.UsePathStyle = true
		})
	}

	s3Client := s3.NewFromConfig(cfg, s3OptFns...)
	logger.Info("S3 FileStorage inicializado", zap.String("region", awsCfg.Region), zap.String("endpoint", awsCfg.Endpoint))

	return &S3FileStorage{s3Client: s3Client, endpoint: awsCfg.Endpoint, region: awsCfg.Region, logger: logger}, nil
}

func (s *S3FileStorage) Upload(ctx context.Context, bucket, key string, reader io.Reader, contentType string) (string, error) {
	input := &s3.PutObjectInput{Bucket: &bucket, Key: &key, Body: reader, ContentType: &contentType}
	_, err := s.s3Client.PutObject(ctx, input)
	if err != nil {
		s.logger.Error("Error subiendo archivo a S3", zap.String("bucket", bucket), zap.String("key", key), zap.Error(err))
		return "", fmt.Errorf("error uploading to S3: %w", err)
	}
	var url string
	if s.endpoint != "" {
		url = fmt.Sprintf("%s/%s/%s", s.endpoint, bucket, key)
	} else {
		url = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, s.region, key)
	}
	return url, nil
}

func (s *S3FileStorage) Delete(ctx context.Context, bucket, key string) error {
	input := &s3.DeleteObjectInput{Bucket: &bucket, Key: &key}
	_, err := s.s3Client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("error deleting from S3: %w", err)
	}
	return nil
}

var _ services.FileStorage = (*S3FileStorage)(nil)
