package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
)

type Database struct {
	Client *redis.Client
	Canvas []byte
}

const CanvasName = "newplaces"

var Ctx = context.TODO()

func NewDatabase(address string) (*Database, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})
	if err := client.Ping(Ctx).Err(); err != nil {
		return nil, err
	}
	bytes, err := client.Get(Ctx, CanvasName).Bytes()
	if err != nil {
		return nil, err
	}
	if len(bytes) != 16_000_000 {
		//for i := 0; i < 4_000_000; i += 4 {
		//	_, err = client.BitField(Ctx, CanvasName, "SET", "u32", 32*i, -1, "SET", "u32", 32*(i+1), -1, "SET", "u32", 32*(i+2), -1, "SET", "u32", 32*(i+3), -1).Result()
		//	if err != nil {
		//		log.Fatal(err.Error())
		//	}
		//	log.Printf("Pixels remaining: %d / 4_000_000", i)
		//}
		return nil, fmt.Errorf("wrong canvas length: %d", len(bytes))
	}
	return &Database{Client: client, Canvas: bytes}, nil
}

func (d *Database) SetPixel(x, y, color uint32) error {
	if x < 0 || x >= 2000 {
		return errors.New("x must be in 0 <= x < 2000 range")
	}
	if y < 0 || y >= 2000 {
		return errors.New("y must be in 0 <= y < 2000 range")
	}
	position := x*2000 + y
	_, err := d.Client.BitField(Ctx, CanvasName, "SET", "u32", 32*position, color).Result()
	if err != nil {
		return err
	}
	position *= 4
	d.Canvas[position] = uint8(color>>24) & 0xFF
	d.Canvas[position+1] = uint8(color>>16) & 0xFF
	d.Canvas[position+2] = uint8(color>>8) & 0xFF
	d.Canvas[position+3] = uint8(color) & 0xFF
	return nil
}
