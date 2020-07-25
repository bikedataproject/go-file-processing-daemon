package main

import (
	"bytes"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/tormoder/fit"
)

func main() {
	// Open .FIT file
	file, err := ioutil.ReadFile("workout.fit")
	if err != nil {
		log.Fatalf("Something went wrong reading the file: ", err)
	}

	// Decode fit file
	fit, err := fit.Decode(bytes.NewReader(file))
	if err != nil {
		log.Fatalf("Could not read .FIT file: %v", err)
	}

	// Print file info
	log.Infof("File created at %v", fit.FileId.TimeCreated)

	// Get activity data
	act, err := fit.Activity()
	if err != nil {
		log.Fatalf("Could not fetch activity: %v", err)
	}

	log.Infof("Activity type: %v", act.Activity.Type)

	// Get positions
	for _, point := range act.Records {
		log.Infof("Activity point: [%v, %v]", point.PositionLat, point.PositionLong)
	}

	for _, session := range act.Sessions {
		log.Infof("Sport type: %v", session.Sport)
	}
}
