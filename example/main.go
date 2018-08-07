package main

import (
	"time"

	"fmt"

	"sync"

	"strings"

	"github.com/oblq/sprbox/example/app"
)

var wg = sync.WaitGroup{}

func main() {
	// Services --------------------------------------------------------------------------------------------------------

	fmt.Printf("This is the primary api URL: %s\n", app.Shared.Services["api"].URL())
	_, alternatives := app.Shared.Services["api"].URLAlternatives()
	fmt.Printf("Those are some alternative api URLs:\n%s\n", strings.Join(alternatives, "\n"))
	fmt.Printf("Those are some custom values stored on api:\n%+v\n\n", app.Shared.Services["api"].Data)

	// Worker-pool -----------------------------------------------------------------------------------------------------

	// calling some funcs on the app.Shared singleton,
	// automatically initialized with the right config.
	jobsNum := 24
	wg.Add(jobsNum)

	i := 0
	for i < jobsNum {
		app.Shared.WP.PushJobAsync(CustomJob{i})
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
