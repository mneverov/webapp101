package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"

	"github.com/mneverov/webapp101/pkg/config"
	"github.com/mneverov/webapp101/pkg/metric"
	"github.com/mneverov/webapp101/pkg/scrape"
)

// Opts contains parsed environment variables and program arguments.
type Opts struct {
	DBOpts        DatabaseOpts
	MigrationOpts MigrationOpts
	AppOpts       ApplicationOpts
}

// DatabaseOpts contains connection parameters for a database.
type DatabaseOpts struct {
	User string `long:"db-user" env:"DB_USER" default:"webapp101" description:"A database user"`
	Pass string `long:"db-pass" env:"DB_PASS" default:"webapp101" description:"A database password"`
	Host string `long:"db-host" env:"DB_HOST" default:"localhost" description:"A database host"`
	Port string `long:"db-port" env:"DB_PORT" default:"5544" description:"A database port"`
	Name string `long:"db-name" env:"DB_NAME" default:"webapp101" description:"A database name"`
}

// MigrationOpts contains options for DB migrations.
type MigrationOpts struct {
	MigrationsDir string `long:"migrations-dir" env:"MIGRATIONS_DIR" default:"resources/migrations" description:"A directory with the DB migration files"`
	MigrateUp     bool   `long:"migrate-up" env:"MIGRATE_UP" description:"Indicates if migrations up should be run on the app start"`
}

// ApplicationOpts contains options for the webapp101 application.
type ApplicationOpts struct {
	Port             int `long:"port" env:"PORT" default:"8080" description:"What port the app should start on"`
	ClientTimeoutSec int `long:"client-timeout-sec" env:"CLIENT_TIMEOUT_SEC" default:"5" description:"Specifies a time limit for requests made by a scraper"`
}

func main() {
	opts, err := parseArguments()
	if err != nil {
		fmt.Printf("failed to parse options %s. Terminating the app\n", err)
		os.Exit(1)
	}

	conn, err := newDBConnection(opts.DBOpts)
	if err != nil {
		fmt.Printf("failed to create DB connection: %s. Terminating the app\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	err = migrate(conn, opts.MigrationOpts)
	if err != nil {
		fmt.Printf("failed to migrate DB: %s. Terminating the app\n", err)
		os.Exit(1)
	}
	metricDB := metric.NewPostgresStorage(conn)
	metricService := metric.NewService(metricDB)
	metricHandler := metric.NewHandler(metricService)

	client := &http.Client{Timeout: time.Duration(opts.AppOpts.ClientTimeoutSec) * time.Second}
	scraperManager := scrape.NewInMemoryManager(client)

	cfgDB := config.NewPostgresStorage(conn)
	cfgService := config.NewService(cfgDB, metricService, scraperManager)
	cfgHandler := config.NewHandler(cfgService)

	router := routes(metricHandler, cfgHandler)
	server := startServer(opts.AppOpts.Port, router)
	stopServerOnSignal(server)
}

func routes(
	metricHandler *metric.Handler, configHandler *config.Handler,
) chi.Router {
	router := chi.NewRouter()
	router.Route("/metrics", func(r chi.Router) {
		r.Get("/", metricHandler.Get)
	})
	router.Route("/configs", func(r chi.Router) {
		r.Get("/", configHandler.GetAll)
		r.Post("/", configHandler.Create)
	})
	return router
}
