package throttler

import "fmt"

type httpPkg struct{}

func (httpPkg) Get(url string) error { return nil }

var http httpPkg

// This example fetches several URLs concurrently,
// using a Throttler to block until all the fetches are complete.
// Compare to http://golang.org/pkg/sync/#example_WaitGroup
func ExampleThrottler() {
	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	// Create a new Throttler that will get 2 urls at a time
	t := New(2, len(urls))
	for _, url := range urls {
		// Launch a goroutine to fetch the URL.
		go func(url string) {
			// Fetch the URL.
			err := http.Get(url)
			// Let Throttler know when the goroutine completes
			// so it can dispatch another worker
			t.Done(err)
		}(url)
		// Pauses until a worker is available or all jobs have been completed
		// Returning the total number of goroutines that have errored
		// lets you choose to break out of the loop without starting any more
		errorCount := t.Throttle()
		if errorCount > 0 {
			break
		}
	}
}

// This example fetches several URLs concurrently,
// using a Throttler to block until all the fetches are complete
// and checks the errors returned.
// Compare to http://golang.org/pkg/sync/#example_WaitGroup
func ExampleThrottler_errors() error {
	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	// Create a new Throttler that will get 2 urls at a time
	t := New(2, len(urls))
	for _, url := range urls {
		// Launch a goroutine to fetch the URL.
		go func(url string) {
			// Let Throttler know when the goroutine completes
			// so it can dispatch another worker
			defer t.Done(nil)
			// Fetch the URL.
			http.Get(url)
		}(url)
		// Pauses until a worker is available or all jobs have been completed
		t.Throttle()
	}

	if t.Err() != nil {
		// Loop through the errors to see the details
		for i, err := range t.Errs() {
			fmt.Printf("error #%d: %s", i, err)
		}
		return t.Err()
	}

	return nil
}
