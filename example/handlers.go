package main

import (
	"net/http"

	"time"

	"github.com/labstack/echo"
)

// Echo ----------------------------------------------------------------------------------------------------------------

func Home(c echo.Context) error {
	app := c.(*CEC).App
	return c.String(http.StatusOK, "Hello, the text in config is: "+app.ATool.GetText())
}

func Text(c echo.Context) error {
	app := c.(*CEC).App
	return c.String(http.StatusOK, "Hello, the text in config is: "+app.ATool2.GetText())
}

// Pool ----------------------------------------------------------------------------------------------------------------

// CustomJob implement the workerful.Job interface
type CustomJob struct {
	ID int
}

// Execute execute the job
func (cj CustomJob) F() error {
	time.Sleep(time.Second)
	println("job", cj.ID, "executed...")
	return nil
}

// DoSomeJobs will use App.workerpool, initialized and configured by sprbox
func DoSomeJobs(c echo.Context) error {
	app := c.(*CEC).App
	jobsNum := 24
	i := 0
	for i < int(jobsNum) {
		app.WPool.PushJobAsync(CustomJob{i})
		i++
	}
	return c.String(http.StatusOK, "Jobs started...")
}
