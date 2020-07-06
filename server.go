package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/noble-constructor/noticoco/routes"
	"github.com/sirupsen/logrus"
	"os"
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	routes.Init(e)

	// start server
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}
