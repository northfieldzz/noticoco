package api

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"net/url"
	"os"
	"path"
)

func FetchCocoVideos() *youtube.SearchListCall {
	key := os.Getenv("API_KEY")
	kiryuCocoId := os.Getenv("COCO_CHANNEL_ID")
	ctx := context.Background()
	yts, err := youtube.NewService(ctx, option.WithAPIKey(key))
	if err != nil {
		logrus.Fatalf("Error creating new Youtube services: %v", err)
	}
	call := yts.Search.List("snippet")
	call = call.ChannelId(kiryuCocoId).Type("video")
	return call
}

func FetchLatestAsacoco() *youtube.SearchResult {
	call := FetchCocoVideos()
	call = call.Q("あさココ").Q("LIVE")
	call = call.Order("date").MaxResults(5)
	res, err := call.Do()
	if err != nil {
		logrus.Fatalf("Error calling Youtube API: %v", err)
	}
	items := res.Items
	return items[0]
}

func GenerateUrl(sr *youtube.SearchResult) *url.URL {
	var endpoint = "https://www.youtube.com"
	u, err := url.Parse(endpoint)
	if err != nil {
		logrus.Fatalf("Error calling url parse: %v", err)
	}
	u.Path = path.Join(u.Path, "watch")
	q := u.Query()
	q.Set("v", sr.Id.VideoId)
	u.RawQuery = q.Encode()
	return u
}
