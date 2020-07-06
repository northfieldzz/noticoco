package routes

import (
	"github.com/labstack/echo"
	"github.com/noble-constructor/noticoco/api"
	"github.com/valyala/fasthttp"
	"os"
)

func Init(e *echo.Echo) {
	{
		e.GET("/", version())
	}
	Line := e.Group("line")
	{
		Line.POST("/asacoco/push", api.Main())
		Line.POST("/asacoco/callback", api.CallBack())
	}
}

func version() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(
			fasthttp.StatusOK,
			map[string]string{"Version": os.Getenv("API_VERSION")},
		)
	}
}
