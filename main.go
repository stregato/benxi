package main

import (
	"flag"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/code-to-go/safepool/api"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/pool"
	"github.com/sirupsen/logrus"
)

var dbName = "safepool.db"

func parseFlags() {
	var verbose int
	var test bool

	flag.IntVar(&verbose, "v", 0, "verbose level - 0 to 2")
	flag.StringVar(&dbName, "d", "", "location of the SQLlite DB")
	flag.BoolVar(&test, "t", false, "in test mode some checks are disable to facilitate development")
	flag.Parse()

	switch verbose {
	case 0:
		logrus.SetLevel(logrus.FatalLevel)
	case 1:
		logrus.SetLevel(logrus.ErrorLevel)
	case 2:
		logrus.SetLevel(logrus.InfoLevel)
	case 3:
		logrus.SetLevel(logrus.DebugLevel)
	}

	if test {
		pool.ForceCreation = true
	}
}

func main() {
	parseFlags()

	dbPath := path.Join(xdg.ConfigHome, dbName)
	err := api.Start(dbPath, pool.HighBandwith)
	if core.IsErr(err, "cannot start: %v") {
		os.Exit(1)
	}
	SelectMain()
}
