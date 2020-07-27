package decode

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"time"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	geo "github.com/paulmach/go.geo"
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
		err = fmt.Errorf("Could not read binary file as .FIT: %v", err)
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
		}
		timestamps = append(timestamps, point.Timestamp)
		unixTimes = append(unixTimes, point.Timestamp.Unix())
	}

	if len(timestamps) < 1 || len(unixTimes) < 1 {
		err = fmt.Errorf("Couldn't convert to contribution: %v", "no data points available")
		return
	}

	// Set contribution values
	contrib.UserAgent = "web/Garmin"
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
