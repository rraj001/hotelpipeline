package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"hotelsdatapipeline/domain"

	"github.com/go-redis/redis/v8"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(host string, port int, db int) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		DB:       db,
		PoolSize: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisRepository{client: client}, nil
}

func (r *RedisRepository) StoreHotelByID(hotelID string, hotel *domain.Hotel) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(hotel)
	if err != nil {
		return fmt.Errorf("failed to marshal hotel: %w", err)
	}

	key := fmt.Sprintf("hotel:id:%s", hotelID)
	if err := r.client.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store hotel: %w", err)
	}

	log.Printf("Stored hotel %s", hotelID)
	return nil
}

func (r *RedisRepository) StoreHotelsByDestinationID(destinationID int, hotels []*domain.Hotel) error {
	if len(hotels) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data, err := json.Marshal(hotels)
	if err != nil {
		return fmt.Errorf("failed to marshal hotels: %w", err)
	}

	key := fmt.Sprintf("hotels:destination:%d", destinationID)
	if err := r.client.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store hotels by destination: %w", err)
	}

	log.Printf("Stored %d hotels for destination %d", len(hotels), destinationID)
	return nil
}

func (r *RedisRepository) GetHotelByID(hotelID string) (*domain.Hotel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("hotel:id:%s", hotelID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("hotel not found: %s", hotelID)
		}
		return nil, fmt.Errorf("failed to get hotel: %w", err)
	}

	var hotel domain.Hotel
	if err := json.Unmarshal(data, &hotel); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hotel: %w", err)
	}

	return &hotel, nil
}

func (r *RedisRepository) GetHotelsByDestinationID(destinationID int) ([]*domain.Hotel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("hotels:destination:%d", destinationID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return []*domain.Hotel{}, nil
		}
		return nil, fmt.Errorf("failed to get hotels by destination: %w", err)
	}

	var hotels []*domain.Hotel
	if err := json.Unmarshal(data, &hotels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hotels: %w", err)
	}

	return hotels, nil
}

func (r *RedisRepository) GetHotelsByIDRange(hotelIDs []string) ([]*domain.Hotel, error) {
	if len(hotelIDs) == 0 {
		return []*domain.Hotel{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, hotelID := range hotelIDs {
		key := fmt.Sprintf("hotel:id:%s", hotelID)
		cmds[hotelID] = pipe.Get(ctx, key)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	var hotels []*domain.Hotel
	for hotelID, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			if err == redis.Nil {
				log.Printf("Hotel not found: %s", hotelID)
				continue
			}
			log.Printf("Failed to get hotel %s: %v", hotelID, err)
			continue
		}

		var hotel domain.Hotel
		if err := json.Unmarshal(data, &hotel); err != nil {
			log.Printf("Failed to unmarshal hotel %s: %v", hotelID, err)
			continue
		}

		hotels = append(hotels, &hotel)
	}

	return hotels, nil
}

func (r *RedisRepository) Close() error {
	return r.client.Close()
}
