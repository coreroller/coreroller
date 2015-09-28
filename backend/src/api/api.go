package api

import (
	"errors"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/mgutz/logxi/v1"
	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/mgutz/dat.v1/sqlx-runner"

	// Postgresql driver
	_ "github.com/lib/pq"
)

const (
	defaultDbURL = "postgres://postgres@127.0.0.1:5432/coreroller?sslmode=disable&connect_timeout=10"
)

var logger = log.New("api")

var (
	// ErrNoRowsAffected indicates that no rows were affected in an update or
	// delete database operation.
	ErrNoRowsAffected = errors.New("coreroller: no rows affected")
)

// API represents an api instance used to interact with CoreRoller entities.
type API struct {
	db       *sqlx.DB
	dbR      *runner.DB
	dbDriver string
	dbURL    string
}

// New creates a new API instance.
func New(options ...func(*API) error) (*API, error) {
	api := &API{
		dbDriver: "postgres",
		dbURL:    os.Getenv("COREROLLER_DB_URL"),
	}
	if api.dbURL == "" {
		api.dbURL = defaultDbURL
	}

	var err error
	api.db, err = sqlx.Open(api.dbDriver, api.dbURL)
	if err != nil {
		return nil, err
	}
	if err := api.db.Ping(); err != nil {
		return nil, err
	}

	dat.EnableInterpolation = true
	api.dbR = runner.NewDBFromSqlx(api.db)

	for _, option := range options {
		err := option(api)
		if err != nil {
			return nil, err
		}
	}

	return api, nil
}

// OptionInitDB will initialize the database during the API instance creation,
// flushing all existing data and loading the schema file. Use with caution.
func OptionInitDB(api *API) error {
	if _, err := api.db.Exec(schemaSQL); err != nil {
		return err
	}
	return nil
}

// Close releases the connections to the database.
func (api *API) Close() {
	_ = api.db.DB.Close()
}
