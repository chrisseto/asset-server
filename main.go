package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go-adapter/gingonic"
	"gopkg.in/gin-gonic/gin.v1"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/sqlite"
)

var (
	databaseFile string
	host         string
	debug        bool
	DB           sqlbuilder.Database
)

func init() {
	flag.StringVar(&databaseFile, "db", "./db/server.db", "The name of the SQLite file.")
	flag.StringVar(&host, "host", "localhost:8080", "The host to bind to")
	flag.BoolVar(&debug, "debug", false, "Run in debug mode?")
}

func setupSync() {
	remotes := map[string]time.Time{}
	for _, remote := range flag.Args() {
		remotes[fmt.Sprintf("http://%s/api/v1/assets", remote)] = time.Unix(0, 0)
	}
	go SyncAssets(remotes)
}

func main() {
	flag.Parse()

	var err error
	DB, err = sqlite.Open(sqlite.ConnectionURL{
		Database: databaseFile,
		Options: map[string]string{
			"cache":         "shared",
			"mode":          "rwc",
			"_busy_timeout": "500000",
		},
	})

	if err != nil {
		panic(err)
	}

	// lolwut
	// 1 connection causes a hang
	// > 2 causes locking issues with any request > 10 ms
	DB.SetMaxOpenConns(2)

	defer DB.Close()

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	api := api2go.NewAPIWithRouting(
		"api/v1",
		api2go.NewStaticResolver("/"),
		gingonic.New(r),
	)

	api.AddResource(Note{}, NewNoteResource())
	api.AddResource(Asset{}, NewAssetResource())

	setupSync()

	r.Run(host)
}
