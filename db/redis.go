package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis(addr, password string, db int) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return RedisClient.Ping(ctx).Err()
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

const (
	DeviceKeyPrefix = "device:"
	DeviceExpire    = 60 * time.Second
)

func RefreshDeviceRedis(deviceID string) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	key := DeviceKeyPrefix + deviceID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return RedisClient.Set(ctx, key, "1", DeviceExpire).Err()
}

func DeleteDeviceRedis(deviceID string) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	key := DeviceKeyPrefix + deviceID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return RedisClient.Del(ctx, key).Err()
}

func GetDeviceRedis(deviceID string) (bool, error) {
	if RedisClient == nil {
		return false, fmt.Errorf("redis client not initialized")
	}
	key := DeviceKeyPrefix + deviceID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
