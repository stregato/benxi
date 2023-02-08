package main

import (
	"flag"
	"path"

	"github.com/adrg/xdg"
	"github.com/sirupsen/logrus"
)

var dbName = "safepool.db"

func parseFlags() {
	var verbose int

	flag.IntVar(&verbose, "v", 0, "verbose level - 0 to 2")
	flag.StringVar(&dbName, "d", "", "location of the SQLlite DB")
	flag.Parse()

	switch verbose {
	case 0:
		logrus.SetLevel(logrus.FatalLevel)
	case 1:
		logrus.SetLevel(logrus.InfoLevel)
	case 2:
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	parseFlags()

	dbPath := path.Join(xdg.ConfigHome, dbName)
	safepool.Start(dbPath)
	SelectMain()
}
