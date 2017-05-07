package api

import (
	"errors"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/rubenv/sql-migrate"
	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/mgutz/dat.v1/sqlx-runner"

	// Postgresql driver
	_ "github.com/lib/pq"
)

//go:generate go-bindata -ignore=\.swp -pkg api db db/migrations

const (
	defaultDbURL = "postgres://postgres@127.0.0.1:5432/coreroller?sslmode=disable&connect_timeout=10"
	nowUTC       = dat.UnsafeString("now() at time zone 'utc'")
)

var (
	// ErrNoRowsAffected indicates that no rows were affected in an update or
	// delete database operation.
	ErrNoRowsAffected = errors.New("coreroller: no rows affected")

	// ErrInvalidSemver indicates that the provided semver version is not valid.
	ErrInvalidSemver = errors.New("coreroller: invalid semver")
)

// API represents an api instance used to interact with CoreRoller entities.
type API struct {
	db       *sqlx.DB
	dbR      *runner.DB
	dbDriver string
	dbURL    string
}

// New creates a new API instance, creating the underlying db connection and
// applying db migrations available.
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

	migrate.SetTable("database_migrations")
	migrations := &migrate.AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "db/migrations",
	}
	if _, err := migrate.Exec(api.db.DB, "postgres", migrations, migrate.Up); err != nil {
		return nil, err
	}

	return api, nil
}

// OptionInitDB will initialize the database during the API instance creation,
// dropping all existing tables, which will force all migration scripts to be
// re-executed. Use with caution, this will DESTROY ALL YOUR DATA.
func OptionInitDB(api *API) error {
	sqlFile, err := Asset("db/drop_all_tables.sql")
	if err != nil {
		return err
	}

	if _, err := api.db.Exec(string(sqlFile)); err != nil {
		return err
	}

	return nil
}

// Close releases the connections to the database.
func (api *API) Close() {
	_ = api.db.DB.Close()
}
