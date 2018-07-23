package main

import (
	"fmt"

	"github.com/labstack/echo"
	"github.com/oblq/sprbox/example/handlers"
	"github.com/xenolf/lego/log"
)

func main() {
	e := echo.New()
	e.HideBanner = true
	e.Use(handlers.EchoSprBox)
	e.GET("/", handlers.Home)
	e.GET("/text", handlers.Text)
	e.GET("/pool", handlers.DoSomeJobs)

	fmt.Println("http://localhost:8888/")
	fmt.Println("http://localhost:8888/text")
	fmt.Println("http://localhost:8888/pool")

	log.Fatal(e.Start(":8888"))
}
