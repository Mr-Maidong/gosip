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

// GetDevicesOnlineStatus 批量检查设备在线状态
// 返回 map[deviceid]bool，true=在线，false=离线
func GetDevicesOnlineStatus(deviceIDs []string) map[string]bool {
	result := make(map[string]bool, len(deviceIDs))
	if RedisClient == nil || len(deviceIDs) == 0 {
		return result
	}

	// 使用 Pipeline 批量检查 key 是否存在
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := RedisClient.Pipeline()
	cmds := make(map[string]*redis.IntCmd, len(deviceIDs))
	for _, deviceID := range deviceIDs {
		cmds[deviceID] = pipe.Exists(ctx, DeviceKeyPrefix+deviceID)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// 出错时全部标记为离线
		for _, deviceID := range deviceIDs {
			result[deviceID] = false
		}
		return result
	}

	// 收集结果
	for deviceID, cmd := range cmds {
		result[deviceID] = cmd.Val() > 0
	}

	return result
}

// ==================== 设备位置缓存 ====================

const (
	GPSKeyPrefix = "gps:position:"
	GPSExpire    = 60 * time.Second
)

// CachedPosition 缓存的位置信息
type CachedPosition struct {
	Longitude float64
	Latitude  float64
	GPSTime   string
	Speed    float64
	Direction float64
	Altitude float64
}

// SetDevicePosition 缓存设备位置
func SetDevicePosition(deviceID string, pos *CachedPosition) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	key := GPSKeyPrefix + deviceID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return RedisClient.HSet(ctx, key, map[string]interface{}{
		"longitude": pos.Longitude,
		"latitude":  pos.Latitude,
		"gps_time":   pos.GPSTime,
		"speed":     pos.Speed,
		"direction": pos.Direction,
		"altitude":  pos.Altitude,
	}).Err()
}

// GetDevicePosition 获取缓存的设备位置
func GetDevicePosition(deviceID string) (*CachedPosition, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	key := GPSKeyPrefix + deviceID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil // 没有缓存
	}
	pos := &CachedPosition{}
	fmt.Sscanf(data["longitude"], "%f", &pos.Longitude)
	fmt.Sscanf(data["latitude"], "%f", &pos.Latitude)
	pos.GPSTime = data["gps_time"]
	fmt.Sscanf(data["speed"], "%f", &pos.Speed)
	fmt.Sscanf(data["direction"], "%f", &pos.Direction)
	fmt.Sscanf(data["altitude"], "%f", &pos.Altitude)
	return pos, nil
}

// GetAllGPSKeys 获取所有GPS缓存key
func GetAllGPSKeys() ([]string, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	iter := RedisClient.Scan(ctx, 0, GPSKeyPrefix+"*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	return keys, iter.Err()
}

// DeleteDevicePosition 删除设备位置缓存
func DeleteDevicePosition(deviceID string) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	key := GPSKeyPrefix + deviceID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return RedisClient.Del(ctx, key).Err()
}
