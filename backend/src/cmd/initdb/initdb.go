package main

import (
	"log"

	"api"
)

func main() {
	if _, err := api.New(api.OptionInitDB); err != nil {
		log.Fatal(err)
	}
}
