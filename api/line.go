package api

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"google.golang.org/api/youtube/v3"
	"net/http"
	"os"
	"regexp"
	"time"
)

const location = "Asia/Tokyo"

var bot, botErr = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))

func CallBack() echo.HandlerFunc {
	return func(c echo.Context) error {
		reply(c.Response().Writer, c.Request())
		return c.JSON(fasthttp.StatusOK, "")
	}
}

func Push() echo.HandlerFunc {
	return func(c echo.Context) error {
		video := FetchLatestAsacoco()
		if isPublishDateOfToday(video.Snippet.Title) {
			if err := pushFlexMessage(video); err != nil {
				logrus.Fatal(err)
			}
		} else {
			if err := pushTextMessage(video); err != nil {
				logrus.Fatal(err)
			}
		}
		return c.JSON(fasthttp.StatusOK, video)
	}
}

func isPublishDateOfToday(target string) bool {
	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, 9*60*60)
	}
	_, todayMonth, todayDate := time.Now().In(loc).Date()
	reg := regexp.MustCompile(fmt.Sprintf("%d月%d日", todayMonth, todayDate))
	return !reg.MatchString(target)
}

func pushTextMessage(video *youtube.SearchResult) error {
	if botErr != nil {
		logrus.Fatalf("linebot Error %v", botErr)
	}
	message := "今日のあさココはお休みです。"
	roomId := os.Getenv("ROOM_ID")
	if _, err := bot.PushMessage(
		roomId,
		linebot.NewTextMessage(message),
	).Do(); err != nil {
		logrus.Fatal(err)
		return err
	}
	return nil
}

func pushFlexMessage(video *youtube.SearchResult) error {
	if botErr != nil {
		logrus.Fatalf("linebot Error %v", botErr)
	}
	container := &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Hero: &linebot.ImageComponent{
			Type:        linebot.FlexComponentTypeImage,
			URL:         video.Snippet.Thumbnails.Medium.Url,
			Size:        linebot.FlexImageSizeTypeFull,
			AspectRatio: linebot.FlexImageAspectRatioType16to9,
			AspectMode:  linebot.FlexImageAspectModeTypeCover,
			Action: &linebot.URIAction{
				Label: "hero01",
				URI:   GenerateUrl(video).String(),
				AltURI: &linebot.URIActionAltURI{
					Desktop: GenerateUrl(video).String(),
				},
			},
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Size:   linebot.FlexTextSizeTypeMd,
					Weight: linebot.FlexTextWeightTypeBold,
					Wrap:   true,
					Text:   video.Snippet.Title,
				},
				&linebot.BoxComponent{
					Type:   linebot.FlexComponentTypeBox,
					Layout: linebot.FlexBoxLayoutTypeBaseline,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type: linebot.FlexComponentTypeText,
							Size: linebot.FlexTextSizeTypeSm,
							Text: video.Snippet.LiveBroadcastContent,
						},
					},
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Height: linebot.FlexButtonHeightTypeSm,
					Style:  linebot.FlexButtonStyleTypeLink,
					Action: &linebot.URIAction{
						Label: video.Snippet.ChannelId,
						URI:   "https://www.youtube.com/channel/UCS9uQI-jC3DE0L4IpXyvr6w",
						AltURI: &linebot.URIActionAltURI{
							Desktop: "https://www.youtube.com/channel/UCS9uQI-jC3DE0L4IpXyvr6w",
						},
					},
				},
				&linebot.ButtonComponent{
					Height: linebot.FlexButtonHeightTypeSm,
					Style:  linebot.FlexButtonStyleTypeLink,
					Action: &linebot.URIAction{
						Label: "ホロライブ公式",
						URI:   "https://www.youtube.com/channel/UCJFZiqLMntJufDCHc6bQixg/featured",
						AltURI: &linebot.URIActionAltURI{
							Desktop: "https://www.youtube.com/channel/UCJFZiqLMntJufDCHc6bQixg/featured",
						},
					},
				},
			},
		},
	}
	roomId := os.Getenv("ROOM_ID")
	if _, err := bot.PushMessage(
		roomId,
		linebot.NewFlexMessage("桐生ココから新しいメッセージです♡", container),
	).Do(); err != nil {
		return err
	}
	return nil
}

func reply(w http.ResponseWriter, req *http.Request) {
	if botErr != nil {
		logrus.Fatal(botErr)
	}
	events, err := bot.ParseRequest(req)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				reply := Researcher(event)
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
					logrus.Print(err)
				}
				//if _, err = bot.ReplyMessage(
				//	event.ReplyToken,
				//	brain(message.Text),
				//).Do(); err != nil {
				//	logrus.Print(err)
				//}
			case *linebot.StickerMessage:
				replyMessage := fmt.Sprintf(
					"sticker id is %s, stickerResourceType is %s",
					message.StickerID,
					message.StickerResourceType,
				)
				if _, err = bot.ReplyMessage(
					event.ReplyToken,
					linebot.NewTextMessage(replyMessage),
				).Do(); err != nil {
					logrus.Print(err)
				}
			}
		}
	}
}

func Researcher(event *linebot.Event) (reply string) {
	message := event.Message.(*linebot.TextMessage).Text
	reply = ""
	var rGroupId = regexp.MustCompile(`#GroupId`)
	if isMatch := rGroupId.MatchString(message); isMatch {
		reply = fmt.Sprintf("GroupId: %s", event.Source.GroupID)
	}
	var rRoomId = regexp.MustCompile(`#RoomId`)
	if isMatch := rRoomId.MatchString(message); isMatch {
		reply = fmt.Sprintf("RoomId: %s", event.Source.RoomID)
	}
	var rUserId = regexp.MustCompile(`#UserId`)
	if isMatch := rUserId.MatchString(message); isMatch {
		reply = fmt.Sprintf("UserId: %s", event.Source.UserID)
	}
	return
}

func brain(message string) *linebot.TextMessage {
	return linebot.NewTextMessage("")
}
