package decode

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	"github.com/google/uuid"
	geo "github.com/paulmach/go.geo"
	"github.com/tkrajina/gpxgo/gpx"
	"github.com/tormoder/fit"
)

// getMax : Fetch the largest item of an array of integers
func getMax(arr []int64) (result int64) {
	result = arr[0]
	for _, item := range arr {
		if item > result {
			result = item
		}
	}
	return
}

// getMin : Fetch the smallest item of an array of integers
func getMin(arr []int64) (result int64) {
	result = arr[0]
	for _, item := range arr {
		if item < result {
			result = item
		}
	}
	return
}

// FitToContribution : Read & decode a FIT file
func FitToContribution(filedir string) (contrib dbmodel.Contribution, err error) {
	// Read file from disk
	file, err := ioutil.ReadFile(filedir)
	if err != nil {
		err = fmt.Errorf("Could not open file %v : %v", filedir, err)
		return
	}

	// Decode binary file
	fit, err := fit.Decode(bytes.NewReader(file))
	if err != nil {
		err = fmt.Errorf("Invalid FIT file format: %v", err)
		os.Remove(filedir)
		return
	}

	// Get activity data
	act, err := fit.Activity()
	if err != nil {
		return
	}

	// Extract path & timestamps from route
	path := geo.NewPath()
	var timestamps []time.Time
	var unixTimes []int64
	for _, point := range act.Records {
		// Add point to complete path & filter out NaN values
		if !math.IsNaN(point.PositionLat.Degrees()) && !math.IsNaN(point.PositionLong.Degrees()) {
			path.Push(geo.NewPoint(point.PositionLong.Degrees(), point.PositionLat.Degrees()))
			timestamps = append(timestamps, point.Timestamp)
			unixTimes = append(unixTimes, point.Timestamp.Unix())
		}
	}

	if len(timestamps) < 1 || len(unixTimes) < 1 {
		err = fmt.Errorf("Couldn't convert to contribution: %v", "no data points available")
		return
	}

	// Set contribution values
	contrib.UserAgent = "web/Fileupload/fit"
	contrib.PointsGeom = path
	contrib.PointsTime = timestamps
	contrib.Distance = int(path.GeoDistance())
	contrib.Duration = int(getMax(unixTimes) - getMin(unixTimes))
	contrib.TimeStampStart = time.Unix(getMin(unixTimes), 0)
	contrib.TimeStampStop = time.Unix(getMax(unixTimes), 0)

	return
}

// GetProviderID : Convert FIT file & get serial number (user provider)
func GetProviderID(filedir string) (userID string, err error) {
	// Read file from disk
	file, err := ioutil.ReadFile(filedir)
	if err != nil {
		err = fmt.Errorf("Could not open file %v : %v", filedir, err)
		return
	}

	// Decode binary file
	fit, err := fit.Decode(bytes.NewReader(file))
	if err != nil {
		err = fmt.Errorf("Could not read binary file as .FIT: %v", err)
		return
	}

	// Set user value
	userID = strconv.FormatUint(uint64(fit.FileId.SerialNumber), 10)
	return
}

// GpxToContribution : Convert a GPX file to contribution
func GpxToContribution(filedir string) (contrib dbmodel.Contribution, err error) {
	// Read file from disk
	file, err := gpx.ParseFile(filedir)
	if err != nil {
		err = fmt.Errorf("Invalid GPX file format: %v", err)
		os.Remove(filedir)
		return
	}

	// Fetch tracks
	path := geo.NewPath()
	var timestamps []time.Time
	var unixTimes []int64
	for _, track := range file.Tracks {
		for _, segment := range track.Segments {
			for _, point := range segment.Points {
				// Set geo point
				if !math.IsNaN(point.Latitude) && !math.IsNaN(point.Longitude) {
					path.Push(geo.NewPoint(point.Longitude, point.Latitude))
					timestamps = append(timestamps, point.Timestamp)
					unixTimes = append(unixTimes, point.Timestamp.Unix())
				}
			}
		}
	}

	if len(timestamps) < 1 || len(unixTimes) < 1 {
		err = fmt.Errorf("Couldn't convert to contribution: %v", "no data points available")
		return
	}

	// Set contribution values
	contrib.UserAgent = "web/Fileupload/gpx"
	contrib.PointsGeom = path
	contrib.PointsTime = timestamps
	contrib.Distance = int(path.GeoDistance())
	contrib.Duration = int(getMax(unixTimes) - getMin(unixTimes))
	contrib.TimeStampStart = time.Unix(getMin(unixTimes), 0)
	contrib.TimeStampStop = time.Unix(getMax(unixTimes), 0)

	return
}

// GetUserFromHTML : extract a user object from an HTML-file
func GetUserFromHTML(filepath string, usr *dbmodel.User) (err error) {
	// Set global user data
	usr.UserIdentifier = uuid.New().String()
	usr.ExpiresAt = -1
	usr.ExpiresIn = -1
	usr.IsHistoryFetched = true
	usr.Provider = "web/LocationHistory"
	usr.TokenCreationDate = time.Now()
	usr.AccessToken = "0"
	usr.RefreshToken = "0"

	// Open HTML file
	file, err := os.Open(filepath)
	if err != nil {
		return
	}

	// Create a buffer reader from the file
	reader := bufio.NewReader(file)

	// Create goquery documentreader
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return
	}

	// Find e-mail address in document
	// Find the header element first
	doc.Find(".header_title").Each(func(i int, s *goquery.Selection) {
		// Split the value of this element by spaces
		pageTitle := strings.Split(s.Text(), " ")
		// Loop over each word
		for _, word := range pageTitle {
			// Find the e-mail address
			if strings.Contains(word, "@") {
				// Hash the e-mail
				hasher := sha512.New()
				hasher.Write([]byte(word))
				usr.ProviderUser = hex.EncodeToString(hasher.Sum(nil))
			}
		}
	})

	return
}
