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

// Push メイン.
func Push() echo.HandlerFunc {
	return func(c echo.Context) error {
		video := FetchLatestVideo()
		if !IsAsacoco(video) || !IsTodayVideo(video) {
			// あさココではない or 今日の動画ではない場合はあさココ休み
			err := PushMessage("今日のあさココはお休みです。")
			if err != nil {
				logrus.Fatal(err)
			}
		} else {
			// あさココの場合はFlexメッセージを送信
			err := pushFlexMessage(video)
			if err != nil {
				logrus.Fatal(err)
			}
		}
		return c.JSON(fasthttp.StatusOK, video)
	}
}

// IsAsacoco 朝ココか.
func IsAsacoco(video *youtube.SearchResult) bool {
	reg := regexp.MustCompile(`あさココLIVE`)
	return reg.MatchString(video.Snippet.Title)
}

// IsTodayVideo 今日の動画か.
func IsTodayVideo(video *youtube.SearchResult) bool {
	time.Local = time.FixedZone("Asia/Tokyo", 9*60*60)
	layout := "2006-01-02T15:04:05Z"
	publishTime, _ := time.Parse(layout, video.Snippet.PublishedAt)
	now := time.Now()
	yesterday := now.Add(time.Duration(-24) * time.Hour)

	if publishTime.Unix() < now.Unix() && publishTime.Unix() > yesterday.Unix() {
		// 過去24時間以内に枠が作成されていればtrueを返却
		return true
	}
	return false
}

// PushMessage メッセージを送信する.
func PushMessage(message string) error {
	if botErr != nil {
		logrus.Fatalf("linebot Error %v", botErr)
	}
	if _, err := bot.PushMessage(
		os.Getenv("ROOM_ID"),
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
				URI:   GenerateURL(video).String(),
				AltURI: &linebot.URIActionAltURI{
					Desktop: GenerateURL(video).String(),
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
							Text: BroadcastStatus(video),
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
						Label: video.Snippet.ChannelTitle,
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
