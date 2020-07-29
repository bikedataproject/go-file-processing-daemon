package main

import (
	"fmt"
	"go-file-processing-daemon/config"
	"go-file-processing-daemon/crawl"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	"github.com/koding/multiconfig"
	log "github.com/sirupsen/logrus"
)

var db dbmodel.Database

// ReadSecret : Read a file and return it's content as string - used for Docker secrets
func ReadSecret(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Could not fetch secret: %v", err)
	}
	return string(data)
}

func main() {
	// Set filetypes
	FileTypes := [2]string{"fit", "gpx"}

	// Set logging to file
	logfile, err := os.OpenFile(fmt.Sprintf("log/%v.log", time.Now().Unix()), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Could not create logfile: %v", err)
	}
	log.SetOutput(logfile)

	// Load configuration values
	conf := &config.Config{}
	multiconfig.MustLoad(conf)

	// Set config
	switch conf.DeploymentType {
	case "production":
		port, _ := strconv.ParseInt(ReadSecret(conf.PostgresPortEnv), 10, 64)
		conf.PostgresHost = ReadSecret(conf.PostgresHost)
		conf.PostgresUser = ReadSecret(conf.PostgresUser)
		conf.PostgresPassword = ReadSecret(conf.PostgresPassword)
		conf.PostgresPort = port
		conf.PostgresDb = ReadSecret(conf.PostgresDb)
		break
	default:
		if conf.PostgresDb == "" || conf.PostgresHost == "" || conf.PostgresPassword == "" || conf.PostgresPort == 0 || conf.PostgresRequireSSL == "" || conf.PostgresUser == "" || conf.FileDir == "" {
			log.Fatal("Configuration not complete")
		}
		break
	}

	// Set database connection
	db = dbmodel.Database{
		PostgresHost:       conf.PostgresHost,
		PostgresUser:       conf.PostgresUser,
		PostgresPassword:   conf.PostgresPassword,
		PostgresPort:       conf.PostgresPort,
		PostgresDb:         conf.PostgresDb,
		PostgresRequireSSL: conf.PostgresRequireSSL,
	}
	db.VerifyConnection()

	// Loop the service forever
	for {
		// Loop over accepted filetypes
		for _, filetype := range FileTypes {
			// Walk through file and handle FIT files
			fitfiles, err := crawl.WalkDirectory(conf.FileDir, filetype)
			if err != nil {
				log.Fatal(err)
			}
			// Process files
			if len(fitfiles) > 0 {
				for _, file := range fitfiles {
					switch filetype {
					case "fit":
						if err := HandleFitFile(file); err != nil {
							log.Errorf("Something went wrong handling a FIT file: %v", err)
						}
						break
					case "gpx":
						if err := HandleGpxFile(file); err != nil {
							log.Errorf("Something went wrong handling a GPX file: %v", err)
						}
						break
					default:
						log.Warnf("Trying to handle a file which is not in filetypes? (%v)", file)
						break
					}
				}
			}
		}
		// Repeat each 10 seconds
		time.Sleep(10 * time.Second)
	}
}
