package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Pixel struct {
	X     uint32 `json:"x"`
	Y     uint32 `json:"y"`
	Color uint32 `json:"color"`
}

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	database, e := NewDatabase("localhost:6379")
	if e != nil {
		log.Fatal(e)
		return nil
	}
	broker := NewBroker()
	go broker.Start()

	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.StaticFile("/", "ui/dist/index.html")
	r.StaticFile("/index.html", "ui/dist/index.html")
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
			case msg := <-messageChannel:
				log.Printf("Message received: %s", msg)
				c.SSEvent("message", msg)
				c.Writer.Flush()
			}
		}
	})

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/
	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	_ = r.Run(":8080")
}
