package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	cr "github.com/coreroller/coreroller/updaters/lib/go"
	"github.com/facebookgo/grace/gracehttp"
)

const (
	checkFrequency = 30 * time.Second
)

var (
	serverAddr     string
	instanceID     string
	appID          = "780d6940-9a48-4414-88df-95ba63bbe9cb"
	groupID        = "51a32aa9-3552-49fc-a28c-6543bccf0069"
	currentVersion = "1.0.0"
)

func init() {
	flag.StringVar(&instanceID, "instance-id", "demo-instance-id", "Instance ID")
}

func main() {
	flag.Parse()

	go checkForUpdates()

	gracehttp.Serve(
		&http.Server{Addr: ":8111", Handler: newHandler("handler1")},
		&http.Server{Addr: ":8112", Handler: newHandler("handler2")},
		&http.Server{Addr: ":8113", Handler: newHandler("handler3")},
	)
}

func checkForUpdates() {
	checkTicker := time.Tick(checkFrequency)

	for range checkTicker {
		log.Println("Checking for updates..")
		update, err := cr.GetUpdate(instanceID, appID, groupID, currentVersion)
		if err != nil {
			log.Printf("\t- No updates (error: %v)\n", err)
			continue
		}
		log.Println("\t- Updates available!")

		log.Println("\t- Downloading update package..")
		cr.EventDownloadStarted(instanceID, appID, groupID)
		downloadPackage(update.URL, update.Filename)

		log.Println("\t- Update package downloaded")
		cr.EventDownloadFinished(instanceID, appID, groupID)

		log.Println("\t- Update completed successfully :)")
		cr.EventUpdateSucceeded(instanceID, appID, groupID)

		log.Println("\t- Restarting server using new package")
		syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)

		break
	}
}

func downloadPackage(url, filename string) {
	time.Sleep(3 * time.Second)

	output, err := os.Create("demo")
	if err != nil {
		return
	}
	defer output.Close()

	response, err := http.Get(url + filename)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if _, err := io.Copy(output, response.Body); err != nil {
		return
	}
}

// From: https://github.com/facebookgo/grace/blob/master/gracedemo/demo.go
func newHandler(name string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sleep/", func(w http.ResponseWriter, r *http.Request) {
		duration, err := time.ParseDuration(r.FormValue("duration"))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		time.Sleep(duration)
		fmt.Fprintf(
			w,
			"(SERVER VERSION: %s) started at %s slept for %f seconds from pid %d.\n",
			currentVersion,
			time.Now(),
			duration.Seconds(),
			os.Getpid(),
		)
	})
	return mux
}

// http://localhost:8111/sleep/?duration=5s
