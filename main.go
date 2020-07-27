package main

import (
	"go-file-processing-daemon/config"
	"go-file-processing-daemon/crawl"
	"go-file-processing-daemon/database"
	"go-file-processing-daemon/decode"
	"time"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	"github.com/google/uuid"
	"github.com/koding/multiconfig"
	log "github.com/sirupsen/logrus"
)

var db database.Database

func main() {
	// Load configuration values
	conf := &config.Config{}
	multiconfig.MustLoad(conf)

	// Set database connection
	db = database.Database{
		PostgresHost:       conf.PostgresHost,
		PostgresUser:       conf.PostgresUser,
		PostgresPassword:   conf.PostgresPassword,
		PostgresPort:       conf.PostgresPort,
		PostgresDb:         conf.PostgresDb,
		PostgresRequireSSL: conf.PostgresRequireSSL,
	}
	db.Connect()

	// Loop the service forever
	for {
		// Walk through file dir
		files, err := crawl.WalkDirectory(conf.FileDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			// Convert FIT to contribution
			contribution, err := decode.FitToContribution(file)
			if err != nil {
				log.Warnf("Could not convert .FIT to contribution: %v", err)
			}

			// Get userID from FIT
			userID, err := decode.GetProviderID(file)
			if err != nil {
				log.Warnf("Could not convert .FIT to user: %v", err)
			}

			// Fetch user data
			user, err := db.GetUserData(userID)
			// Check if user exists; if not create a new object
			if user.ID == "" {
				user = dbmodel.User{
					Provider:          "web/Garmin",
					ProviderUser:      userID,
					IsHistoryFetched:  true,
					ExpiresAt:         -1,
					ExpiresIn:         -1,
					TokenCreationDate: time.Now(),
					UserIdentifier:    uuid.New().String(),
				}
				usr, err := db.AddUser(&user)
				if err != nil {
					log.Fatalf("Could not create new user: %v", err)
				}
				user = usr
			}

			// Add contribution
			if err := db.AddContribution(&contribution, &user); err != nil {
				log.Errorf("Could not create contribution: %v", err)
			} else {
				log.Infof("Added contribution for user %v", userID)
			}
		}

		// Repeat each minute
		time.Sleep(1 * time.Minute)
	}
}
