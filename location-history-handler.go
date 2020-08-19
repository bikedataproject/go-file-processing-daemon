package main

import (
	"archive/zip"
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	"github.com/google/uuid"
	geo "github.com/paulmach/go.geo"
	log "github.com/sirupsen/logrus"
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
func HandleLocationFile(filepath string, user dbmodel.User) error {
	// Attempt to read the file
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	// Unmarshal file
	var history LocationHistory
	if err = json.Unmarshal(data, &history); err != nil {
		return fmt.Errorf("Could not unmarshall data into location history: %v", err)
	}

	if len(history.Locations) > 0 {
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
					if act.Type == LocationHistoryCylcingType && act.Confidence >= LocationHistoryActivityThreshold {
						// Set trip
						trips[timestamp.Format("2006-01-02")] = append(trips[timestamp.Format("2006-01-02")], point)
						break
					}
				}
			}
		}

		// Convert map to Contributions
		contributions, err := tripsToContributions(trips)
		if err != nil {
			return err
		}

		// Upload data to database
		for _, contribution := range contributions {
			if err := db.AddContribution(&contribution, &user); err != nil {
				log.Warnf("Could not add contribution to database: %v", err)
			} else {
				log.Info("Add location history trip to database")
			}
		}

	} else {
		return fmt.Errorf("%v is not a location history file or is empty", filepath)
	}

	// Clean return
	return nil
}

// UnpackLocationFiles : Unzip a given .ZIP file's contents
func UnpackLocationFiles(filepath string, extractPath string) (locationfiles []string, foldername string, err error) {
	// Unzip & get all filenames
	foldername = fmt.Sprintf("%v/%v", extractPath, uuid.New())
	files, err := unzip(filepath, foldername)
	if err != nil {
		return
	}

	// Search for the location history files
	for _, file := range files {
		if strings.Contains(file, ".json") || strings.Contains(file, ".html") {
			locationfiles = append(locationfiles, file)
		}
	}
	return
}

// tripsToContributions : Convert location history trips to bikedataproject Contributions
func tripsToContributions(trips map[string][]LocationHistoryPoint) (contributions []dbmodel.Contribution, err error) {
	for _, trip := range trips {
		// Check if trip contains more points then the threshold
		if len(trip) >= LocationHistoryPointThreshold {
			// Create geopath from points
			geoPath := geo.NewPath()
			var timestamps []time.Time

			for _, point := range trip {
				// Add geopoint to path
				geoPath.Push(geo.NewPoint(point.LongitudeE7/1e7, point.LatitudeE7/1e7))

				// Get point timestamp
				unixMs, err := strconv.ParseInt(point.TimestampMs, 10, 64)
				if err != nil {
					return contributions, err
				}
				// Convert to UNIX timestamp
				ts := time.Unix(unixMs/1000, 0)
				timestamps = append(timestamps, ts)
			}

			// Create contribution
			contrib := dbmodel.Contribution{
				UserAgent:      "web/LocationHistory",
				TimeStampStart: getStartTimestamp(trip),
				TimeStampStop:  getEndTimestamp(trip),
				Distance:       int(geoPath.GeoDistance()),
				Duration:       int(getEndTimestamp(trip).Sub(getEndTimestamp(trip)).Seconds()),
				PointsGeom:     geoPath,
				PointsTime:     timestamps,
			}

			// Add contribution to array
			contributions = append(contributions, contrib)
		}
	}
	return
}

// getStartTimestamp : get the lowest timestamp of an array of LocationHistoryPoints
func getStartTimestamp(points []LocationHistoryPoint) (timestamp time.Time) {
	// Set timestamp to now
	timestamp = time.Now()

	// Loop over trip points
	for _, p := range points {
		if tmpTimestamp, err := getTimestamp(p); err == nil {
			// Check if timestamp is earlier
			if diff := timestamp.Sub(tmpTimestamp); diff > 0 {
				timestamp = tmpTimestamp
			}
		}
	}
	return
}

// getStartTimestamp : get the highest timestamp of an array of LocationHistoryPoints
func getEndTimestamp(points []LocationHistoryPoint) (timestamp time.Time) {
	// Loop over trip points
	for _, p := range points {
		if tmpTimestamp, err := getTimestamp(p); err == nil {
			// Check if timestamp is earlier
			if diff := timestamp.Sub(tmpTimestamp); diff < 0 {
				timestamp = tmpTimestamp
			}
		}
	}
	return
}

// getTimestamp : Get the timestamp of a single LocationHistoryPoint
func getTimestamp(point LocationHistoryPoint) (timestamp time.Time, err error) {
	unixMs, err := strconv.ParseInt(point.TimestampMs, 10, 64)
	if err != nil {
		return
	}

	// Convert to UNIX timestamp
	timestamp = time.Unix(unixMs/1000, 0)
	return
}

// getUserProvider : Read HTML-file to fetch provider user
func getProviderUser(filepath string) (id string, err error) {
	// Read file
	file, err := os.Open(filepath)
	if err != nil {
		return
	}

	// Convert to bytesreader
	reader := bufio.NewReader(file)

	// Convert to Goquery object
	doc, err := goquery.NewDocumentFromReader(reader)

	// Find page title & Loop over results - should be just 1
	doc.Find(".header_title").Each(func(i int, s *goquery.Selection) {
		// Split sentence in words
		pageTitle := strings.Split(s.Text(), " ")
		for _, word := range pageTitle {
			// Extract e-mail address
			if strings.Contains(word, "@") {
				// Hash with SHA1
				hasher := sha1.New()
				hasher.Write([]byte(word))
				id = hex.EncodeToString(hasher.Sum(nil))
				break
			}
		}
	})

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
