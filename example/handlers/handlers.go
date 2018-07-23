package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/oblq/sprbox/example/toolbox"
)

// CEC is a custom echo context
type CEC struct {
	echo.Context
	App *toolbox.ToolBox
}

// EchoSprBox provides the AppBox (inherited from echo.Context) to echo.
// This middleware should be registered before any other.
func EchoSprBox(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// pass the pointer to 'app'
		embeddedBox := &CEC{Context: c, App: &toolbox.App}
		return h(embeddedBox)
	}
}

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

// F execute execute the job
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
	for i < jobsNum {
		app.WPool.PushJobAsync(CustomJob{i})
		i++
	}
	return c.String(http.StatusOK, "Jobs started...")
}
