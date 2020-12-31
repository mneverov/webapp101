package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fortytw2/dockertest"
	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
	"github.com/go-testfixtures/testfixtures/v3"
)

// StartPostgresContainer starts postgres container with given options. It
// returns a representation of a container with the Address (host:port) of the
// newly started container.
func StartPostgresContainer(opts pg.Options) *dockertest.Container {
	waitFunc := func(addr string) error {
		opts.Addr = addr
		conn := pg.Connect(&opts)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return conn.Ping(ctx)
	}

	container, err := dockertest.RunContainer(
		"postgres:13.1-alpine",
		strings.Split(opts.Addr, ":")[1],
		waitFunc,
		"-e", fmt.Sprintf("POSTGRES_DB=%s", opts.Database),
		"-e", fmt.Sprintf("POSTGRES_USER=%s", opts.User),
		"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", opts.Password),
		"--rm")
	if err != nil {
		panic(err)
	}

	return container
}

// TestDB prepares a DB for testing:
// - It runs migrate UP before test suite execution and migrate DOWN on the test
// cleanup.
// - It applies fixtures to the newly created tables.
// - It returns a connection to the DB, note that the DB should be already
// running.
func TestDB(t *testing.T, dbOpts pg.Options, suiteName string) *pg.DB {
	conn := pg.Connect(&dbOpts)
	t.Cleanup(func() {
		_ = conn.Close()
	})

	err := conn.Ping(context.Background())
	if err != nil {
		_ = conn.Close()
		t.Fatalf("%s: failed to ping DB %+v", suiteName, err)
	}

	t.Logf(
		"%s: connected to \"postgres://%s:***@%s/%s\" sslmode enabled=%t\n",
		suiteName, dbOpts.User, dbOpts.Addr, dbOpts.Database,
		dbOpts.TLSConfig != nil,
	)

	migrate(t, conn, suiteName)

	fixturesConn, err := connectDB(&dbOpts)
	if err != nil {
		t.Fatalf("%s: failed to connect %+v\n", suiteName, err)
	}
	err = loadFixtures(fixturesConn)
	if err != nil {
		t.Fatalf("%s: failed to load fixtures %+v\n", suiteName, err)
	}

	return conn
}

func migrate(t *testing.T, migrationsDB *pg.DB, suiteName string) {
	col := migrations.NewCollection()
	err := col.DiscoverSQLMigrations("../../resources/migrations")
	if err != nil {
		t.Fatalf("%s failed to discover migrations %+v\n", suiteName, err)
	}

	// it is mandatory to run init before migrations:
	// see https://github.com/go-pg/migrations/issues/48
	_, _, err = col.Run(migrationsDB, "init")
	if err != nil {
		t.Fatalf("%s failed to init migrations %+v\n", suiteName, err)
	}

	// even if Run accepts vararg of commands, all should run separately
	oldVersion, newVersion, err := col.Run(migrationsDB, "up")
	if err != nil {
		t.Fatalf("%s: failed to apply migrations %+v\n", suiteName, err)
	}
	t.Logf(
		"%s: apply migrations from %d to %d \n",
		suiteName, oldVersion, newVersion,
	)

	t.Cleanup(func() {
		oldVersion, newVersion, err = col.Run(migrationsDB, "reset")
		if err != nil {
			t.Fatalf("%s: failed to reset migrations %+v\n", suiteName, err)
		}
		t.Logf(
			"%s: reset migrations from %d to %d \n",
			suiteName, oldVersion, newVersion,
		)
	})
}

func loadFixtures(conn *sql.DB) error {
	fixtures, err := testfixtures.New(
		testfixtures.Database(conn),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("../../resources/fixtures"),
	)
	if err != nil {
		return err
	}

	return fixtures.Load()
}

func connectDB(opts *pg.Options) (*sql.DB, error) {
	source := fmt.Sprintf(
		"user=%s password='%s' host=%s port=%s dbname=%s sslmode=disable",
		opts.User,
		opts.Password,
		strings.Split(opts.Addr, ":")[0],
		strings.Split(opts.Addr, ":")[1],
		opts.Database,
	)
	conn, err := sql.Open("postgres", source)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
