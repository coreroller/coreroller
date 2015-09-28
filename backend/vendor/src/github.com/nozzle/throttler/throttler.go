// Throttler fills the gap between sync.WaitGroup and manually monitoring your goroutines
// with channels. The API is almost identical to Wait Groups, but it allows you to set
// a max number of workers that can be running simultaneously. It uses channels internally
// to block until a job completes by calling Done(err) or until all jobs have been completed.
//
// After exiting the loop where you are using Throttler, you can call the `Err` or `Errs` method to check
// for errors. `Err` will return a single error representative of all the errors Throttler caught. The
// `Errs` method will return all the errors as a slice of errors (`[]error`).
//
// Compare the Throttler example to the sync.WaitGroup example http://golang.org/pkg/sync/#example_WaitGroup
//
// See a fully functional example on the playground at http://bit.ly/throttler-v3
package throttler

import (
	"fmt"
	"math"
	"sync"
)

type Throttler struct {
	maxWorkers    int
	workerCount   int
	batchingTotal int
	batchSize     int
	totalJobs     int
	jobsStarted   int
	jobsCompleted int
	doneChan      chan struct{}
	errsMutex     *sync.Mutex
	errs          []error
	errorCount    int
}

// New returns a Throttler that will govern the max number of workers and will
// work with the total number of jobs. It panics if maxWorkers < 1.
func New(maxWorkers, totalJobs int) *Throttler {
	if maxWorkers < 1 {
		panic("maxWorkers has to be at least 1")
	}
	return &Throttler{
		maxWorkers: maxWorkers,
		batchSize:  1,
		totalJobs:  totalJobs,
		doneChan:   make(chan struct{}, totalJobs),
		errsMutex:  &sync.Mutex{},
	}
}

// NewBatchedThrottler returns a Throttler (just like New), but also enables batching.
func NewBatchedThrottler(maxWorkers, batchingTotal, batchSize int) *Throttler {
	totalJobs := int(math.Ceil(float64(batchingTotal) / float64(batchSize)))
	t := New(maxWorkers, totalJobs)
	t.batchSize = batchSize
	t.batchingTotal = batchingTotal
	return t
}

// Throttle works similarly to sync.WaitGroup, except inside your goroutine dispatch
// loop rather than after. It will not block until the number of active workers
// matches the max number of workers designated in the call to NewThrottler or
// all of the jobs have been dispatched. It stops blocking when Done has been called
// as many times as totalJobs.
func (t *Throttler) Throttle() int {
	if t.totalJobs < 1 {
		return t.errorCount
	}
	t.jobsStarted++
	t.workerCount++

	if t.workerCount == t.maxWorkers {
		<-t.doneChan
		t.jobsCompleted++
		t.workerCount--
	}

	if t.jobsStarted == t.totalJobs {
		for t.jobsCompleted < t.totalJobs {
			<-t.doneChan
			t.jobsCompleted++
		}
	}

	return t.errorCount
}

// Done lets Throttler know that a job has been completed so that another worker
// can be activated. If Done is called less times than totalJobs,
// Throttle will block forever
func (t *Throttler) Done(err error) {
	t.doneChan <- struct{}{}
	if err != nil {
		t.errsMutex.Lock()
		t.errs = append(t.errs, err)
		t.errorCount++
		t.errsMutex.Unlock()
	}
}

// Err returns an error representative of all errors caught by throttler
func (t *Throttler) Err() error {
	if len(t.errs) == 0 {
		return nil
	}
	return multiError(t.errs)
}

// Errs returns a slice of any errors that were received from calling Done()
func (t *Throttler) Errs() []error {
	return t.errs
}

type multiError []error

func (te multiError) Error() string {
	errString := te[0].Error()
	if len(te) > 1 {
		errString += fmt.Sprintf(" (and %d more errors)", len(te)-1)
	}
	return errString
}

func (t *Throttler) BatchStartIndex() int {
	return t.jobsStarted * t.batchSize
}

func (t *Throttler) BatchEndIndex() int {
	end := (t.jobsStarted + 1) * t.batchSize
	if end > t.batchingTotal {
		end = t.batchingTotal
	}
	return end
}

func (t *Throttler) TotalJobs() int {
	return t.totalJobs
}
