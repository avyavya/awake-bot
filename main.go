package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/carlescere/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	bot, err := linebot.New(os.Getenv("LINE_CHANNEL_SECRET"), os.Getenv("LINE_CHANNEL_TOKEN"))

	if err != nil {
		log.Fatal(err)
	}

	job := func() {
		t := time.Now()
		log.Println("[info] Time's up! @", t.UTC())
	}

	// todo JST
	scheduler.Every().Day().At("22:01").Run(job)

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	// Setup HTTP Server for receiving requests from LINE platform
	router.POST("/message", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				c.Writer.WriteHeader(400)
			} else {
				c.Writer.WriteHeader(500)
			}
			return
		}

		for _, event := range events {
			log.Printf("[info] event type: %s", event.Type)
			log.Printf("[info] user id: %s", event.Source.UserID)

			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					log.Printf("[info] message: %s", message.Text)
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})

	// for UptimeRobot
	router.GET("/ping", onPing)
	// will be ignored all this below
	router.Run(":" + port)
}

func onPing(c *gin.Context) {
	c.String(http.StatusOK, "")
}
