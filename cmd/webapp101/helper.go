package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

func startServer(port int, router http.Handler) *http.Server {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("starting webapp101 service on port %d\n", port)
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("an error occurred after stopping the server %s\n", err)
		}
	}()

	return server
}

func stopServerOnSignal(server *http.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	sig := <-sigCh

	log.Printf("shutdown webapp101 service due to received signal %q\n", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	err := server.Shutdown(ctx)
	cancel()

	if err != nil {
		log.Printf("failed to shut down webapp101 service: %s\n", err)
	}
}

func newDBConnection(opts DatabaseOpts) (*pg.DB, error) {
	pgOpts := pg.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		User:     opts.User,
		Password: opts.Pass,
		Database: opts.Name,
	}
	conn := pg.Connect(&pgOpts)
	err := conn.Ping(context.Background())
	if err != nil {
		_ = conn.Close()
		return nil, errors.Wrapf(
			err, "failed to ping DB %s, addr %s, user %s",
			pgOpts.Database, pgOpts.Addr, pgOpts.User,
		)
	}

	log.Printf(
		"connected to \"postgres://%s:***@%s/%s\"\n",
		pgOpts.User, pgOpts.Addr, pgOpts.Database,
	)

	return conn, nil
}

func migrate(conn *pg.DB, migrationOpts MigrationOpts) error {
	if !migrationOpts.MigrateUp {
		return nil
	}

	col := migrations.NewCollection()
	err := col.DiscoverSQLMigrations(migrationOpts.MigrationsDir)
	if err != nil {
		return errors.Wrapf(
			err, "failed to discover migrations in %s", migrationOpts.MigrationsDir,
		)
	}

	_, _, err = col.Run(conn, "init")
	if err != nil {
		return errors.Wrapf(err, "failed to init migration")
	}

	oldVersion, newVersion, err := col.Run(conn, "up")
	if err != nil {
		return errors.Wrapf(err, "failed to apply migrations")
	}
	log.Printf("applied migrations from %d to %d\n", oldVersion, newVersion)

	return nil
}

// parseArguments parses the application arguments and environment variables.
func parseArguments() (*Opts, error) {
	var opts Opts
	p := flags.NewParser(&opts, flags.Default)

	if _, err := p.Parse(); err != nil {
		return nil, fmt.Errorf("error when parsing flags: %s", err)
	}

	return &opts, nil
}
