package main

import (
	"fmt"

	"github.com/labstack/echo"
	"github.com/xenolf/lego/log"
)

func main() {
	e := echo.New()
	e.HideBanner = true
	e.Use(EchoSprBox)
	e.GET("/", home)
	e.GET("/text", text)
	e.GET("/pool", doSomeJobs)

	fmt.Println("http://localhost:8888/")
	fmt.Println("http://localhost:8888/text")
	fmt.Println("http://localhost:8888/pool")

	log.Fatal(e.Start(":8888"))
}
