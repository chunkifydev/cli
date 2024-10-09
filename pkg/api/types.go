package api

import (
	"fmt"
	"sort"
	"strings"
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
	Id        string   `json:"id"`
	Url       string   `json:"url"`
	ProjectId string   `json:"project_id"`
	Enabled   bool     `json:"enabled"`
	Events    []string `json:"events"`
}

type WebhookWithSecretKey struct {
	Webhook
	SecretKey string `json:"secret_key"`
}

type Function struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	ProjectId string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	Enabled   bool      `json:"enabled"`
	Events    []string  `json:"events"`
}

type FunctionInvoked struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
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

type WebhookPayload struct {
	Event string             `json:"event"`
	Date  time.Time          `json:"date"`
	Data  WebhookPayloadData `json:"data"`
}

type WebhookPayloadData struct {
	JobId    string  `json:"job_id"`
	Metadata any     `json:"metadata"`
	SourceId string  `json:"source_id"`
	Error    *string `json:"error"`
	Files    []File  `json:"files"`
}

type Log struct {
	Time       time.Time `json:"time"`
	Level      string    `json:"level"`
	Msg        string    `json:"msg"`
	Service    string    `json:"service"`
	Attributes LogAttrs  `json:"attributes"`
	JobId      string    `json:"job_id,omitempty"`
}

func (l Log) AttributesString() string {
	if l.Attributes == nil {
		return ""
	}
	return l.Attributes.String()
}

type LogAttrs map[string]any

func (a LogAttrs) String() string {
	attrs := []string{}
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if a[k] != nil {
			attrs = append(attrs, fmt.Sprintf("%s=%v", k, a[k]))
		}
	}

	return strings.Join(attrs, " ")
}
