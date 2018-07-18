package main

import (
	"net/http"

	"time"

	"github.com/labstack/echo"
)

// Echo ----------------------------------------------------------------------------------------------------------------

func home(c echo.Context) error {
	app := c.(*CEC).App
	return c.String(http.StatusOK, "Hello, the text in config is: "+app.ATool.getText())
}

func text(c echo.Context) error {
	app := c.(*CEC).App
	return c.String(http.StatusOK, "Hello, the text in config is: "+app.ATool2.getText())
}

// Pool ----------------------------------------------------------------------------------------------------------------

// CustomJob implement the workerful.Job interface
type CustomJob struct {
	ID int
}

// F execute execute the job
func (cj CustomJob) F() error {
	time.Sleep(time.Second)
	println("job", cj.ID, "executed...")
	return nil
}

// DoSomeJobs will use App.workerpool, initialized and configured by sprbox
func doSomeJobs(c echo.Context) error {
	app := c.(*CEC).App
	jobsNum := 24
	i := 0
	for i < jobsNum {
		app.WPool.PushJobAsync(CustomJob{i})
		i++
	}
	return c.String(http.StatusOK, "Jobs started...")
}
