package backend

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// Database holds the redis client and the canvas cached for new requests
type Database struct {
	// Client from redis server
	Client redis.Cmdable
	// Canvas cached from redis, so we don't need to get it for every new client
	Canvas []byte
}

const (
	CanvasName   = "newplaces"
	CanvasWidth  = 2_000
	CanvasHeight = 2_000
	TotalPixels  = CanvasWidth * CanvasHeight
	TotalBytes   = TotalPixels * 4
)

// NewDatabase checks the connection with redis and caches the canvas with prior length checks.
func NewDatabase(client redis.Cmdable) (*Database, error) {
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	bytes, err := client.Get(ctx, CanvasName).Bytes()
	if err != nil {
		return nil, err
	}
	if len(bytes) != TotalBytes {
		return nil, fmt.Errorf("wrong canvas length: %d", len(bytes))
	}
	return &Database{Client: client, Canvas: bytes}, nil
}

// SetPixel sets the pixel in either redis(persistence) and Canvas(mem cache)
func (d *Database) SetPixel(x, y, color uint32) error {
	if x < 0 || x >= CanvasWidth {
		return errors.New("x must be in 0 <= x < 2000 range")
	}
	if y < 0 || y >= CanvasHeight {
		return errors.New("y must be in 0 <= y < 2000 range")
	}
	position := y*CanvasHeight + x
	if _, err := d.Client.BitField(ctx, CanvasName, "SET", "u32", 32*position, color).Result(); err != nil {
		return err
	}
	position *= 4
	d.Canvas[position] = uint8(color>>24) & 0xFF
	d.Canvas[position+1] = uint8(color>>16) & 0xFF
	d.Canvas[position+2] = uint8(color>>8) & 0xFF
	d.Canvas[position+3] = uint8(color) & 0xFF
	return nil
}
