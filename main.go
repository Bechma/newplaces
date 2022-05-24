package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type Pixel struct {
	X     uint32 `json:"x"`
	Y     uint32 `json:"y"`
	Color uint32 `json:"color"`
}

var palette = []uint32{
	0xFFFFFFFF, // white
	0xE4E4E4FF, // light grey
	0x888888FF, // grey
	0x222222FF, // black
	0xFFA7D1FF, // pink
	0xE50000FF, // red
	0xE59500FF, // orange
	0xA06A42FF, // brown
	0xE5D900FF, // yellow
	0x94E044FF, // lime
	0x02BE01FF, // green
	0x00D3DDFF, // cyan
	0x0083C7FF, // blue
	0x0000EAFF, // dark blue
	0xCF6EE4FF, // magenta
	0x820080FF, // purple
}

func setupRouter() *gin.Engine {
	database, e := NewDatabase("127.0.0.1:6379")
	if e != nil {
		log.Fatal(e.Error())
		return nil
	}
	broker := NewBroker()
	go broker.Start()

	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.StaticFile("/", "ui/dist/index.html")
	r.Static("/assets", "ui/dist/assets")

	r.GET("/canvas", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/octet-stream", database.Canvas)
	})
	r.POST("/pixel", func(c *gin.Context) {
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad color"})
		}
		if err := database.SetPixel(px.X, px.Y, px.Color); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		broker.Publish(px)
		c.String(http.StatusOK, "OK")
	})

	r.GET("/events", func(c *gin.Context) {
		messageChannel := broker.Subscribe()
		defer close(messageChannel)
		defer broker.Unsubscribe(messageChannel)
		clientGone := c.Writer.CloseNotify()
		// Handshake to fire up the connect event in the browser
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
					log.Printf("Message received: %s", msg)
					c.SSEvent("message", msg)
					c.Writer.Flush()
				} else {
					return
				}
			}
		}
	})

	r.GET("/palette", func(c *gin.Context) {
		c.JSON(http.StatusOK, palette)
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err.Error())
	}
}
