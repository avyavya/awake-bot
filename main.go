package main

import (
	"awake-bot/timeout"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

const (
	AwakeBotTokenEnv = "AWAKE_BOT_TOKEN"
)

var (
	// LineBot Client
	bot      *linebot.Client
	snooze   *timeout.Timeout
	repeated int
)

func init() {
	tz := os.Getenv("TZ")
	if tz == "" {
		return
	}

	local, _ := time.LoadLocation(tz)
	log.Println("Timezone: " + local.String())
	time.Local = local
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	lb, err := linebot.New(os.Getenv("LINE_CHANNEL_SECRET"), os.Getenv("LINE_CHANNEL_TOKEN"))

	if err != nil {
		log.Fatal(err)
	}

	bot = lb

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	// Setup HTTP Server for receiving requests from LINE platform
	router.POST("/message", onMessage)
	// push message via HTTP request
	router.POST("/push", onPush)
	// for UptimeRobot
	router.GET("/ping", onPing)
	// pushTest()

	// will be ignored all this below
	router.Run(":" + port)
}

func pushTest() {
	bot.PushMessage("C377079ced8ae010da2a12f5e2e365f30", linebot.NewTextMessage("test!")).Do()
}

func onPing(c *gin.Context) {
	c.Writer.WriteHeader(http.StatusOK)
}

// when message received from LINE
func onMessage(c *gin.Context) {
	events, err := bot.ParseRequest(c.Request)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			c.Writer.WriteHeader(http.StatusBadRequest)
		} else {
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, event := range events {
		log.Printf("[info] event type: %s", event.Type)
		log.Printf("[info] user id: %s, group id: %s", event.Source.UserID, event.Source.GroupID)

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
}

// when received a push-message via webhook
func onPush(c *gin.Context) {
	token := os.Getenv(AwakeBotTokenEnv)

	if token != c.PostForm("token") {
		c.Writer.WriteHeader(http.StatusNotFound)
		log.Printf("[err] Invalid token %s", c.PostForm("token"))
		return
	}

	roomId := c.PostForm("room_id")

	if roomId == "" {
		log.Printf("[err] 'room_id' is missing.")
		return
	}

	log.Printf("[info] push message target id: %s", roomId)

	message := c.PostForm("message")

	if message == "" {
		log.Printf("[err] 'message' is missing.")
		return
	}

	wait, _ := strconv.Atoi(c.DefaultPostForm("timeout", "0"))

	if wait > 0 {
		snooze = timeout.New(onTimeout, wait, roomId)
		repeated = 0
	}

	if err := pushMessage(roomId, message); err != nil {
		log.Print(err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Printf("[info] message pushed.")
		c.Writer.WriteHeader(http.StatusOK)
	}
}

func pushMessage(roomId string, message string) error {
	_, err := bot.PushMessage(roomId, linebot.NewTextMessage(message)).Do()
	return err
}

func onTimeout(roomId string, wait int) {
	if repeated < 5 {
		msg := "おーい。起きてるかー？？"
		pushMessage(roomId, msg)

		repeated++
		log.Printf("[info] snooze %d with timeout: %d sec for roomId %s", repeated, wait, roomId)
		snooze = timeout.New(onTimeout, wait, roomId)
	} else {
		log.Printf("[info] snooze repeated %d times.", repeated)
	}
}
