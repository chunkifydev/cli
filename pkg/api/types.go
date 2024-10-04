package api

import (
	"encoding/json"
	"time"
)

type Source struct {
	Id             string         `json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	Url            string         `json:"url"`
	Metadata       map[string]any `json:"metadata"`
	Device         string         `json:"device"`
	Size           int64          `json:"size"`
	Duration       int64          `json:"duration"`
	Width          int64          `json:"width"`
	Height         int64          `json:"height"`
	VideoBitrate   int64          `json:"video_bitrate"`
	VideoCodec     string         `json:"video_codec"`
	VideoFramerate float64        `json:"video_framerate"`
	AudioCodec     string         `json:"audio_codec"`
	AudioBitrate   int64          `json:"audio_bitrate"`
	Images         []string       `json:"images"`
	Jobs           []Job          `json:"jobs"`
}

type Job struct {
	Id           string     `json:"id"`
	Status       string     `json:"status"`
	Progress     float64    `json:"progress"`
	BillableTime int64      `json:"billable_time"`
	SourceId     string     `json:"source_id"`
	Storage      JobStorage `json:"storage"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	StartedAt    time.Time  `json:"started_at"`
	Transcoder   Transcoder `json:"transcoder"`
	Template     Template   `json:"template"`
	Metadata     any        `json:"metadata"`
}

type Template struct {
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Config  map[string]any `json:"config"`
}

type Transcoder struct {
	Type     string   `json:"type"`
	Quantity int64    `json:"quantity"`
	Speed    float64  `json:"speed"`
	Status   []string `json:"status"`
}

type JobStorage struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Path   string `json:"path"`
}

type File struct {
	Id        string    `json:"id"`
	JobId     string    `json:"job_id"`
	Storage   string    `json:"storage"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
	Url       string    `json:"url,omitempty"`
}

type Project struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Paused    bool      `json:"paused"`
	Storage   string    `json:"storage"`
}

type Storage struct {
	Provider string `json:"provider"`
	Bucket   string `json:"bucket"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	Public   bool   `json:"public"`
	Reserved bool   `json:"reserved"`
}

type Webhook struct {
	Id        string `json:"id"`
	Url       string `json:"url"`
	ProjectId string `json:"project_id"`
	Enabled   bool   `json:"enabled"`
	Events    string `json:"events"`
}

type WebhookWithSecretKey struct {
	Webhook
	SecretKey string `json:"secret_key"`
}

type Function struct {
	Id          string    `json:"id"`
	Description string    `json:"description"`
	ProjectId   string    `json:"project_id"`
	CreatedAt   time.Time `json:"created_at"`
	Enabled     bool      `json:"enabled"`
	Events      string    `json:"events"`
}

type Notification struct {
	Id                 string    `json:"id"`
	JobId              string    `json:"job_id"`
	Type               string    `json:"type"`
	CreatedAt          time.Time `json:"created_at"`
	Attempts           int       `json:"attempts"`
	Payload            string    `json:"payload"`
	ResponseStatusCode int       `json:"response_status_code"`
	Event              string    `json:"event"`
}

type Log struct {
	Time     time.Time  `json:"time"`
	Level    string     `json:"level"`
	Msg      string     `json:"msg"`
	Service  string     `json:"service"`
	LogAttrs S3LogAttrs `json:"-"`
}

type S3LogAttrs map[string]any

// Custom UnmarshalJSON to capture dynamic fields
func (l *Log) UnmarshalJSON(data []byte) error {
	type Alias Log // Create an alias to avoid recursive calls
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal dynamic fields separately
	var rawMap map[string]any
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	// Remove known fields to leave only dynamic ones
	delete(rawMap, "time")
	delete(rawMap, "level")
	delete(rawMap, "msg")
	delete(rawMap, "service")

	l.LogAttrs = rawMap
	return nil
}

func (l *Log) MarshalJSON() ([]byte, error) {
	// Fixed fields
	type Alias Log
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	}

	// Marshal fixed fields
	fixedFields, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	// Marshal dynamic fields
	dynamicFields, err := json.Marshal(l.LogAttrs)
	if err != nil {
		return nil, err
	}

	// Merge fixed and dynamic fields into one map
	var fixedMap map[string]any
	if err := json.Unmarshal(fixedFields, &fixedMap); err != nil {
		return nil, err
	}
	var dynamicMap map[string]any
	if err := json.Unmarshal(dynamicFields, &dynamicMap); err != nil {
		return nil, err
	}

	// Combine maps
	for key, value := range dynamicMap {
		fixedMap[key] = value
	}

	// Return combined JSON
	return json.Marshal(fixedMap)
}
