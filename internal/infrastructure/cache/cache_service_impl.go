// internal/infrastructure/cache/cache_service_impl.go
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisCacheService struct {
	client *RedisClient
	logger *zap.Logger
}

func NewRedisCacheService(client *RedisClient, logger *zap.Logger) services.CacheService {
	return &RedisCacheService{client: client, logger: logger}
}

func (s *RedisCacheService) Get(ctx context.Context, key string) (string, error) {
	val, err := s.client.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		s.logger.Warn("Error obteniendo cache", zap.String("key", key), zap.Error(err))
		return "", err
	}
	return val, nil
}

func (s *RedisCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Client.Set(ctx, key, string(data), ttl).Err()
}

func (s *RedisCacheService) Delete(ctx context.Context, key string) error {
	return s.client.Client.Del(ctx, key).Err()
}

func (s *RedisCacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := s.client.Client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := s.client.Client.Del(ctx, iter.Val()).Err(); err != nil {
			s.logger.Warn("Error eliminando cache por patrón", zap.String("key", iter.Val()), zap.Error(err))
		}
	}
	return iter.Err()
}

var _ services.CacheService = (*RedisCacheService)(nil)
