package throttler

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestThrottle(t *testing.T) {
	var tests = []struct {
		Desc       string
		Jobs       []string
		MaxWorkers int
		TotalJobs  int
	}{
		{
			"Standard implementation",
			[]string{"job01", "job02", "job03", "job04", "job05", "job06", "job07", "job08", "job09", "job10",
				"job11", "job12", "job13", "job14", "job15", "job16", "job17", "job18", "job19", "job20",
				"job21", "job22", "job23", "job24", "job25", "job26", "job27", "job28", "job29", "job30",
				"job31", "job32", "job33", "job34", "job35", "job36", "job37", "job38", "job39", "job40",
				"job41", "job42", "job43", "job44", "job45", "job46", "job47", "job48", "job49", "job50"},
			5,
			-1,
		}, {
			"Incorrectly has 0 as TotalWorkers",
			[]string{"job01", "job02", "job03", "job04", "job05", "job06", "job07", "job08", "job09", "job10",
				"job11", "job12", "job13", "job14", "job15", "job16", "job17", "job18", "job19", "job20",
				"job21", "job22", "job23", "job24", "job25", "job26", "job27", "job28", "job29", "job30",
				"job31", "job32", "job33", "job34", "job35", "job36", "job37", "job38", "job39", "job40",
				"job41", "job42", "job43", "job44", "job45", "job46", "job47", "job48", "job49", "job50"},
			5,
			0,
		}, {
			"More workers than jobs",
			[]string{"job01", "job02", "job03", "job04", "job05", "job06", "job07", "job08", "job09", "job10",
				"job11", "job12", "job13", "job14", "job15", "job16", "job17", "job18", "job19", "job20",
				"job21", "job22", "job23", "job24", "job25", "job26", "job27", "job28", "job29", "job30",
				"job31", "job32", "job33", "job34", "job35", "job36", "job37", "job38", "job39", "job40",
				"job41", "job42", "job43", "job44", "job45", "job46", "job47", "job48", "job49", "job50"},
			50000,
			-1,
		},
	}

	for _, test := range tests {
		totalJobs := len(test.Jobs)
		if test.TotalJobs != -1 {
			totalJobs = test.TotalJobs
		}
		th := New(test.MaxWorkers, totalJobs)
		for _, job := range test.Jobs {
			go func(job string, th *Throttler) {
				defer th.Done(nil)
				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
			}(job, th)
			th.Throttle()
		}
		if th.Err() != nil {
			fmt.Println("err:", th.Err())
		}
	}
}

func TestThrottleWithErrors(t *testing.T) {
	var tests = []struct {
		Desc       string
		Jobs       []string
		MaxWorkers int
		TotalJobs  int
	}{
		{
			"Standard implementation",
			[]string{"job01", "job02", "job03", "job04", "job05", "job06", "job07", "job08", "job09", "job10",
				"job11", "job12", "job13", "job14", "job15", "job16", "job17", "job18", "job19", "job20",
				"job21", "job22", "job23", "job24", "job25", "job26", "job27", "job28", "job29", "job30",
				"job31", "job32", "job33", "job34", "job35", "job36", "job37", "job38", "job39", "job40",
				"job41", "job42", "job43", "job44", "job45", "job46", "job47", "job48", "job49", "job50"},
			5,
			-1,
		}, {
			"Standard implementation",
			[]string{"job01", "job02"},
			5,
			-1,
		},
	}

	for _, test := range tests {
		totalJobs := len(test.Jobs)
		if test.TotalJobs != -1 {
			totalJobs = test.TotalJobs
		}
		th := New(test.MaxWorkers, totalJobs)
		for _, job := range test.Jobs {
			go func(job string, th *Throttler) {
				jobNum, _ := strconv.ParseInt(job[len(job)-2:], 10, 8)
				var err error
				if jobNum%2 != 0 {
					err = fmt.Errorf("Error on %s", job)
				}
				defer th.Done(err)

				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
			}(job, th)
			th.Throttle()
		}
		if len(th.Errs()) != totalJobs/2 {
			t.Fatal("The wrong number of errors were returned")
		}
		if th.Err() != nil {
			fmt.Println("err:", th.Err())
		}
	}
}

func TestThrottlePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Test failed to panic")
		}
	}()
	New(0, 100)
}

func TestBatchedThrottler(t *testing.T) {
	var tests = []struct {
		Desc                  string
		ToBeBatched           []string
		MaxWorkers            int
		BatchSize             int
		ExpectedBatchedSlices [][]string
	}{
		{
			"Standard implementation",
			[]string{"item01", "item02", "item03", "item04", "item05", "item06", "item07", "item08", "item09", "item10",
				"item11", "item12", "item13", "item14", "item15", "item16", "item17", "item18", "item19", "item20",
				"item21", "item22", "item23", "item24", "item25", "item26", "item27", "item28", "item29", "item30",
				"item31", "item32", "item33", "item34", "item35", "item36", "item37", "item38", "item39", "item40",
				"item41", "item42", "item43", "item44", "item45", "item46", "item47", "item48", "item49",
			},
			10,
			2,
			[][]string{
				{"item01", "item02"},
				{"item03", "item04"},
				{"item05", "item06"},
				{"item07", "item08"},
				{"item09", "item10"},
				{"item11", "item12"},
				{"item13", "item14"},
				{"item15", "item16"},
				{"item17", "item18"},
				{"item19", "item20"},
				{"item21", "item22"},
				{"item23", "item24"},
				{"item25", "item26"},
				{"item27", "item28"},
				{"item29", "item30"},
				{"item31", "item32"},
				{"item33", "item34"},
				{"item35", "item36"},
				{"item37", "item38"},
				{"item39", "item40"},
				{"item41", "item42"},
				{"item43", "item44"},
				{"item45", "item46"},
				{"item47", "item48"},
				{"item49"},
			},
		},
	}

	for _, test := range tests {
		th := NewBatchedThrottler(test.MaxWorkers, len(test.ToBeBatched), test.BatchSize)
		for i := 0; i < th.TotalJobs(); i++ {
			go func(tbbSlice []string, expectedSlice []string) {
				if !reflect.DeepEqual(tbbSlice, expectedSlice) {
					t.Fatalf("wanted: %#v | got: %#v", expectedSlice, tbbSlice)
				}
				th.Done(nil)
			}(test.ToBeBatched[th.BatchStartIndex():th.BatchEndIndex()], test.ExpectedBatchedSlices[i])
			if errCount := th.Throttle(); errCount > 0 {
				break
			}
		}

		if th.Err() != nil {
			fmt.Println("err:", th.Err())
		}
	}
}
