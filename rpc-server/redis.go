package main

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	cli *redis.Client
}

func (c *RedisClient) InitClient(ctx context.Context, address, password string) error {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	// test Redis connection
	if err := r.Ping(ctx).Err(); err != nil {
		return err
	}

	c.cli = r
	return nil
}

func (c *RedisClient) SaveMessage(ctx context.Context, roomID string, message *Message) error {
	// Marshal the Go struct into JSON
	text, err := json.Marshal(message)
	if err != nil {
		return err
	}

	member := &redis.Z{
		Score:  float64(message.Timestamp), // The sort key
		Member: text,                       // Data
	}

	// insert
	_, err = c.cli.ZAdd(ctx, roomID, *member).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *RedisClient) GetMessagesByRoomID(ctx context.Context, roomID string, start, end int64, getReverse bool) ([]*Message, error) {
	var (
		rawMessages []string
		messages    []*Message
		err         error
	)

	if getReverse {
		// Desc order -> first message is the latest message
		rawMessages, err = c.cli.ZRevRange(ctx, roomID, start, end).Result()
		if err != nil {
			return nil, err
		}
	} else {
		// Asc order -> first message is the earliest message
		rawMessages, err = c.cli.ZRange(ctx, roomID, start, end).Result()
		if err != nil {
			return nil, err
		}
	}

	for _, msg := range rawMessages {
		temp := &Message{}
		// Unmarshal back into struct
		err := json.Unmarshal([]byte(msg), temp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, temp)
	}

	return messages, nil
}
