package main

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

// ReadLocationFile : Parse a given JSON file and process it's contents
func HandleLocationFile(filepath string) error {
	return nil
}

// UnpackLocationFiles : Unzip a given .ZIP file's contents
func UnpackLocationFiles(filepath string) error {
	return nil
}
