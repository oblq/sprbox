package main

import (
	"time"

	"fmt"

	"sync"

	"github.com/oblq/sprbox"
	"github.com/oblq/sprbox/example/app"
	"github.com/oblq/workerful"
)

var wg = sync.WaitGroup{}

func main() {
	// manually loading configuration on an omitted tool: --------------------------------------------------------------

	var cfg workerful.Config
	if err := sprbox.LoadConfig(&cfg, "./config/WPool"); err != nil {
		println(err.Error())
	}
	app.Shared.WPoolOmitted.Workerful = *workerful.New("", &cfg)
	app.Shared.WPoolOmitted.PushFunc(func() error {
		println("printed from a job in worker-pool...")
		return nil
	})

	// using tools: ----------------------------------------------------------------------------------------------------

	// calling some funcs on the app.Shared singleton,
	// automatically initialized with the right config.
	jobsNum := 24
	wg.Add(jobsNum)

	i := 0
	for i < jobsNum {
		app.Shared.WPool.PushJobAsync(CustomJob{i})
		i++
	}

	wg.Wait()
}

// Define a customJob to use in our worker-pool ------------------------------------------------------------------------

// CustomJob implement the workerful.Job interface
type CustomJob struct {
	ID int
}

// F execute execute the job
func (cj CustomJob) F() error {
	time.Sleep(3 * time.Second)
	fmt.Println("job", cj.ID, "executed...")
	wg.Done()
	return nil
}
