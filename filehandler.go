package main

import (
	"fmt"
	"go-file-processing-daemon/decode"
	"os"
	"time"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// HandleFitFile : Process a single fit file
func HandleFitFile(file string) error {
	// Convert FIT to contribution
	contribution, err := decode.FitToContribution(file)
	if err != nil {
		return fmt.Errorf("Could not convert .FIT to contribution: %v", err)
	}

	// Get userID from FIT
	userID, err := decode.GetProviderID(file)
	if err != nil {
		return fmt.Errorf("Could not convert .FIT to user: %v", err)
	}

	// Fetch user data
	user, err := db.GetUserData(userID)
	// Check if user exists; if not create a new object
	if user.ID == "" {
		user = dbmodel.User{
			Provider:          "web/Fileupload",
			ProviderUser:      userID,
			IsHistoryFetched:  true,
			ExpiresAt:         -1,
			ExpiresIn:         -1,
			TokenCreationDate: time.Now(),
			UserIdentifier:    uuid.New().String(),
		}
		usr, err := db.AddUser(&user)
		if err != nil {
			return fmt.Errorf("Could not create new user: %v", err)
		}
		user = usr
	}

	// Add contribution
	if err := db.AddContribution(&contribution, &user); err != nil {
		return fmt.Errorf("Could not create contribution: %v", err)
	}
	log.Infof("Added contribution for user %v", userID)
	os.Remove(file)
	return nil
}

// HandleGpxFile : Process a single GPX file
func HandleGpxFile(file string) error {
	contribution, err := decode.GpxToContribution(file)
	if err != nil {
		return fmt.Errorf("Could not convert .FIT to contribution: %v", err)
	}

	// Use anonymous user since GPX files
	// Fetch user data
	user, err := db.GetUserData("AnonymousFileUpload")
	// Check if user exists; if not create a new object
	if user.ID == "" {
		user = dbmodel.User{
			Provider:          "web/Fileupload",
			ProviderUser:      "AnonymousFileUpload",
			IsHistoryFetched:  true,
			ExpiresAt:         -1,
			ExpiresIn:         -1,
			TokenCreationDate: time.Now(),
			UserIdentifier:    uuid.New().String(),
		}
		usr, err := db.AddUser(&user)
		if err != nil {
			return fmt.Errorf("Could not create new user: %v", err)
		}
		user = usr
	}

	// Add contribution
	if err := db.AddContribution(&contribution, &user); err != nil {
		return fmt.Errorf("Could not create contribution: %v", err)
	}
	log.Infof("Added contribution for user %v", user.ID)
	os.Remove(file)
	return nil
}
