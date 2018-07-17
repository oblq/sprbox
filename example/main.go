package main

import (
	"fmt"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	e.HideBanner = true
	e.Use(EchoSprBox)
	e.GET("/", Home)
	e.GET("/text", Text)
	e.GET("/pool", DoSomeJobs)

	fmt.Println("http://localhost:8888/")
	fmt.Println("http://localhost:8888/text")
	fmt.Println("http://localhost:8888/pool")

	e.Start(":8888")
}
