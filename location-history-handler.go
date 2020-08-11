package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	log "github.com/sirupsen/logrus"
)

const (
	// BikeCertaintyThreshold : Threshold to validate the activity confidence against
	BikeCertaintyThreshold = 40
	// BikeType : Type of activity which matches bike riding
	BikeType = "ON_BICYCLE"
)

// PointActivity : Single activity information object
type PointActivity struct {
	Type       string `json:"type"`
	Confidence int    `json:"confidence"`
}

// PointActivities : Collection of activities
type PointActivities struct {
	TimestampMs string          `json:"timestampMs"`
	Activity    []PointActivity `json:"activity"`
}

// LocationHistoryPoint : Single location datapoint
type LocationHistoryPoint struct {
	TimestampMs string            `json:"timestampMs"`
	LatitudeE7  float64           `json:"latitudeE7"`
	LongitudeE7 float64           `json:"longitudeE7"`
	Accuracy    int               `json:"accuracy"`
	Activity    []PointActivities `json:"activity,omitempty"`
}

// LocationHistory : Collection of LocationHistoryPoints
type LocationHistory struct {
	Locations []LocationHistoryPoint `json:"locations"`
}

// HandleLocationFile : Parse a given JSON file and process it's contents
func HandleLocationFile(filepath string) error {
	// Attempt to read the file
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	// Unmarshal file
	var history LocationHistory
	if err = json.Unmarshal(data, &history); err != nil {
		return err
	}

	// Convert history to trip-based objects
	trips := make(map[string][]LocationHistoryPoint)

	// Loop over each individual point & organise per day
	for _, point := range history.Locations {
		unixMs, err := strconv.ParseInt(point.TimestampMs, 10, 64)
		if err != nil {
			return err
		}
		timestamp := time.Unix(unixMs/1000, 0)

		// Loop over the activities for each point
		for _, actCollection := range point.Activity {
			for _, act := range actCollection.Activity {
				if act.Type == BikeType && act.Confidence >= BikeCertaintyThreshold {
					// Set trip
					trips[timestamp.Format("2006-01-02")] = append(trips[timestamp.Format("2006-01-02")], point)
					break
				}
			}
		}
	}

	// Clean return
	log.Info(filepath)
	return nil
}

// UnpackLocationFiles : Unzip a given .ZIP file's contents
func UnpackLocationFiles(filepath string) (locationfiles []string, err error) {
	// Unzip & get all filenames
	files, err := unzip(filepath, fmt.Sprintf(""))
	if err != nil {
		log.Fatal(err)
	}

	// Search for the location history files
	for _, file := range files {
		if strings.Contains(file, ".json") {
			locationfiles = append(locationfiles, file)
		}
	}
	return
}

// tripsToContributions : Convert location history trips to bikedataproject Contributions
func tripsToContributions(trips map[string][]LocationHistoryPoint) (contributions []dbmodel.Contribution, err error) {
	for _, trip := range trips {
		tsStart, err := getStartTimestamp(trip)
		if err != nil {
			return contributions, err
		}
		tsStop, err := getEndTimestamp(trip)
		if err != nil {
			return contributions, err
		}

		_ = dbmodel.Contribution{
			UserAgent:      "web/LocationHistory",
			TimeStampStart: time.Unix(tsStart, 0),
			TimeStampStop:  time.Unix(tsStop, 0),
			Distance:       0,
			Duration:       0,
		}
	}

	return
}

func getStartTimestamp(points []LocationHistoryPoint) (timestamp int64, err error) {
	// Set timestamp to now
	timestamp = time.Now().Unix()

	// Loop over trip points
	for _, p := range points {
		// Get timestamp in milliseconds
		unixMs, err := strconv.ParseInt(p.TimestampMs, 10, 64)
		if err != nil {
			return 0, err
		}
		// Convert to UNIX timestamp
		unix := unixMs / 1000
		// Check if timestamp is earlier
		if unix < timestamp {
			timestamp = unix
		}
	}
	return
}

func getEndTimestamp(points []LocationHistoryPoint) (timestamp int64, err error) {
	// Set timestamp to 1970
	timestamp = 0

	// Loop over trip points
	for _, p := range points {
		// Get timestamp in milliseconds
		unixMs, err := strconv.ParseInt(p.TimestampMs, 10, 64)
		if err != nil {
			return 0, err
		}
		// Convert to UNIX timestamp
		unix := unixMs / 1000
		// Check if timestamp is earlier
		if unix > timestamp {
			timestamp = unix
		}
	}
	return
}

// unzip : unzip a given .zip file and return the filenames of the contents
func unzip(source string, destination string) (result []string, err error) {
	var filenames []string

	reader, err := zip.OpenReader(source)
	if err != nil {
		return filenames, err
	}
	defer reader.Close()

	for _, f := range reader.File {
		// Store filename/path for returning and using later on
		path := filepath.Join(destination, f.Name)

		// Add filename to result
		result = append(result, path)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(path, os.ModePerm)
			continue
		}

		// Copy file & contents
		if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return result, err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return result, err
		}

		rc, err := f.Open()
		if err != nil {
			return result, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return
}
