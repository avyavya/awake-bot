package main

import (
	"awake-bot/timeout"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pinzolo/flagday"
)

const (
	AwakeBotTokenEnv = "AWAKE_BOT_TOKEN"
)

var (
	bot    *linebot.Client             // LineBot Client
	snooze map[string]*timeout.Timeout // roomId
)

func init() {
	tz := os.Getenv("TZ")
	if tz == "" {
		return
	}

	local, _ := time.LoadLocation(tz)
	log.Println("Timezone: " + local.String())
	time.Local = local

	snooze = map[string]*timeout.Timeout{}
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
	router.HEAD("/ping", onPing)
	router.GET("/ping", onPing)

	// will be ignored all this below
	router.Run(":" + port)
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

				if message.Text == "/id" {
					reply := "UserId: " + event.Source.UserID + ", GroupId: " + event.Source.GroupID
					bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do()
					return
				}

				if to, e := snooze[event.Source.GroupID]; e {
					if event.Source.UserID == to.GetMonitoringUserId() {
						r := regexp.MustCompile(`^おはよ.`)
						if r.MatchString(message.Text) {
							log.Printf("[info] monitoring user id %s is matched. stop monitoring.", to.GetMonitoringUserId())

							msg := "おはよー！！\n今日も一日がんばるぞい☀"
							bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do()
							// ReplyToken available for 1 time only
							pushSticker(to.RoomId, "3", "193")

							to.Stop()
							delete(snooze, to.RoomId)
							return
						}
					}
				}
			}
		}
	}
}

// when received a push-message via webhook
func onPush(c *gin.Context) {
	token := os.Getenv(AwakeBotTokenEnv)

	if isHolidayToday() {
		log.Printf("[info] Today is holiday. // todo skip")
	}

	if token != c.PostForm("token") {
		c.Writer.WriteHeader(http.StatusNotFound)
		log.Printf("[err] Invalid token %s", c.PostForm("token"))
		return
	}

	userId := c.PostForm("user_id")

	if userId == "" {
		c.Writer.WriteHeader(http.StatusBadRequest)
		log.Printf("[err] 'user_id' is missing.")
		return
	}

	log.Printf("[info] monitoring target user id: %s", userId)

	roomId := c.PostForm("room_id")

	if roomId == "" {
		roomId = userId
	}

	log.Printf("[info] push message target id: %s", roomId)

	message := c.PostForm("message")

	if message == "" {
		log.Printf("[err] 'message' is missing.")
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	wait, _ := strconv.Atoi(c.DefaultPostForm("timeout", "0"))

	if wait > 0 {
		if _, exists := snooze[roomId]; exists {
			log.Printf("[err] snooze for roomId %s is already exists.", roomId)
			return
		} else {
			snooze[roomId] = timeout.New(onTimeout, wait, roomId, userId)
		}
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

func pushSticker(roomId string, packageId string, stickerId string) error {
	msg := linebot.NewStickerMessage(packageId, stickerId)
	_, err := bot.PushMessage(roomId, msg).Do()
	return err
}

func onTimeout(to *timeout.Timeout) {
	if to.Repeated < 5 { // todo
		msg := "おーい。起きてるかー？？"
		pushMessage(to.RoomId, msg)
		pushSticker(to.RoomId, "11537", "52002744")

		log.Printf("[info] snooze %d with timeout: %d sec for roomId %s", to.Repeated, to.Sec, to.RoomId)
		to.Snooze()
	} else {
		pushSticker(to.RoomId, "11537", "52002764")
		log.Printf("[info] snooze repeated %d times.", to.Repeated)
		delete(snooze, to.RoomId)
	}
}

func isHolidayToday() bool {
	today := time.Now()
	return today.Weekday() == 0 || today.Weekday() == 6 || flagday.IsPublicHolidayTime(today)
}
