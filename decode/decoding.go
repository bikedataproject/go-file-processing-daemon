package decode

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"time"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	geo "github.com/paulmach/go.geo"
	"github.com/tormoder/fit"
)

func getMax(arr []int64) (result int64) {
	result = arr[0]
	for _, item := range arr {
		if item > result {
			result = item
		}
	}
	return
}

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
func FitToContribution(filedir string) (result dbmodel.Contribution, err error) {
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

	// Set values
	result.UserAgent = "web/Garmin"
	result.PointsGeom = path
	result.PointsTime = timestamps
	result.Distance = int(path.GeoDistance())
	result.Duration = int(getMax(unixTimes) - getMin(unixTimes))
	result.TimeStampStart = time.Unix(getMin(unixTimes), 0)
	result.TimeStampStop = time.Unix(getMax(unixTimes), 0)

	return
}
