package main

import (
	"go-file-processing-daemon/config"
	"go-file-processing-daemon/crawl"
	"go-file-processing-daemon/decode"

	"github.com/koding/multiconfig"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Load configuration values
	conf := &config.Config{}
	multiconfig.MustLoad(conf)

	// Walk through file dir
	files, err := crawl.WalkDirectory(conf.FileDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		f, err := decode.FitToContribution(file)
		if err != nil {
			log.Warnf("Could not convert .FIT to contribution: %v", err)
		}
		log.Info(f.TimeStampStart)
	}
}
