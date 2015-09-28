package api

import (
	"log"
	"os"
	"testing"
)

const (
	testsDbURL string = "postgres://postgres@127.0.0.1:5432/coreroller_tests?sslmode=disable&connect_timeout=10"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("COREROLLER_DB_URL", testsDbURL)

	a, err := New(OptionInitDB)
	if err != nil {
		log.Println("These tests require PostgreSQL running and a tests database created, please adjust testsDbUrl as needed.")
		log.Println("Default: postgres://postgres@127.0.0.1:5432/coreroller_tests?sslmode=disable&connect_timeout=10")
		os.Exit(1)
	}
	a.Close()

	os.Exit(m.Run())
}
