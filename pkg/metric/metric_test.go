package metric

import (
	"os"
	"testing"

	"github.com/go-pg/pg/v10"

	"github.com/mneverov/webapp101/pkg/testutil"
)

var dbOpts pg.Options

func TestMain(m *testing.M) {
	opts := pg.Options{
		Addr:     "127.0.0.1:5432",
		User:     "webapp101",
		Password: "webapp101",
		Database: "webapp101_test",
	}
	os.Exit(func() int {
		container := testutil.StartPostgresContainer(opts)
		opts.Addr = container.Addr
		dbOpts = opts
		defer container.Shutdown()
		return m.Run()
	}())
}
