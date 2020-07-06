# noticoco


Virtual Youtuber 桐生ココの動向を知らせるLine bot 

## development

```shell script
$ go version
go version go1.14.2

$ go mod download
$ go run server.go
```

## deployment

### Environment
- API_KEY  
GCP youtube data apiの認証APIキー

- API_VERSION  
このAPIのバージョン指定

- PORT  
リスニングポート指定

- COCO_CHANNEL_ID  
桐生ココ チャンネルID

- LINE_REPLY_URL  
Line Webhook URL

- CHANNEL_SECRET  
Line Webhook Secret Key

- CHANNEL_TOKEN  
Line Webhook Token

- ROOM_ID  
メッセージをプッシュするline id

## copyright

This source code includes the following license

### [BSD-3-Clause](https://opensource.org/licenses/BSD-3-Clause)

- Copyright (c) 2011 Google Inc. All rights reserved.  
https://github.com/googleapis/google-api-go-client

### [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0)

- Copyright (C) 2016 LINE Corp.  
LINE Messaging API SDK for Go  
https://github.com/line/line-bot-sdk-go

### MIT License
- Copyright (c) 2017 LabStack  
Echo  
https://github.com/labstack/echo

- Copyright (c) 2014 Simon Eskildsen  
Logrus    
https://github.com/sirupsen/logrus

- Copyright (c) 2015-2016 Aliaksandr Valialkin, VertaMedia  
fasthttp  
https://github.com/valyala/fasthttp

