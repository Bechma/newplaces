package backend

import (
	"log"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func SetupRouter(redisClient redis.Cmdable) (*gin.Engine, error) {
	database, err := NewDatabase(redisClient)
	if err != nil {
		return nil, err
	}
	broker := NewBroker()
	go broker.Start()

	r := gin.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.StaticFile("/", "ui/dist/index.html")
	r.Static("/assets", "ui/dist/assets")

	r.GET("/canvas", getCanvas(database))
	r.POST("/pixel", setPixel(database, broker))
	r.GET("/events", sendEvents(broker))
	r.GET("/palette", getPalette())

	return r, nil
}

func getCanvas(db *Database) func(*gin.Context) {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "application/octet-stream", db.Canvas)
	}
}

func getPalette() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, palette)
	}
}

func setPixel(db *Database, broker *Broker) func(*gin.Context) {
	return func(c *gin.Context) {
		var px Pixel
		if err := c.ShouldBindJSON(&px); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		found := false
		for _, val := range palette {
			if val == px.Color {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad color value"})
			return
		}
		if err := db.SetPixel(px.X, px.Y, px.Color); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		broker.Publish(px)
		c.String(http.StatusOK, "OK")
	}
}

func sendEvents(broker *Broker) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		messageChannel := broker.Subscribe()
		defer close(messageChannel)
		defer broker.Unsubscribe(messageChannel)
		clientGone := c.Writer.CloseNotify()
		// Handshake to fire up the connect event in the browser
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Flush()
		for {
			select {
			case <-clientGone:
				log.Println("Client disconnected")
				return
			case msg, ok := <-messageChannel:
				if ok {
					c.SSEvent("message", msg)
					c.Writer.Flush()
				} else {
					return
				}
			}
		}
	}
}
