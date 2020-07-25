package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tormoder/fit"
)

// Point : example data struct
type Point struct {
	Lat       float64   `json:"lat"`
	Long      float64   `json:"lon"`
	Timestamp time.Time `json:"timestamp"`
	Speed     uint16
}

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

	var points []Point

	// Get positions
	for _, point := range act.Records {
		p := Point{
			Lat:       point.PositionLat.Degrees(),
			Long:      point.PositionLong.Degrees(),
			Timestamp: point.Timestamp,
			Speed:     point.Speed,
		}
		points = append(points, p)
	}

	// Convert to JSON
	pointsJSON, err := json.Marshal(points)
	if err != nil {
		log.Fatalf("Could not convert points to JSON: %v", err)
	} else {
		log.Info(string(pointsJSON))
	}

	for _, session := range act.Sessions {
		log.Infof("Sport type: %v", session.Sport)
	}
}
