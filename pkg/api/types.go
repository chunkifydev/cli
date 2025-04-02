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
	Id            string     `json:"id"`
	Status        string     `json:"status"`
	Progress      float64    `json:"progress"`
	BillableTime  int64      `json:"billable_time"`
	SourceId      string     `json:"source_id"`
	HlsManifestId *string    `json:"hls_manifest_id"`
	Storage       JobStorage `json:"storage"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	StartedAt     time.Time  `json:"started_at"`
	Transcoder    Transcoder `json:"transcoder"`
	Template      Template   `json:"template"`
	Metadata      any        `json:"metadata"`
}

type Template struct {
	Name   string         `json:"name"`
	Config map[string]any `json:"config"`
}

type Transcoder struct {
	Type     string   `json:"type"`
	Quantity int64    `json:"quantity"`
	Speed    float64  `json:"speed"`
	Status   []string `json:"status"`
}

type JobStorage struct {
	Id     string `json:"id"`
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
	Id       string `json:"id"`
	Provider string `json:"provider"`
	Bucket   string `json:"bucket"`
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

type Token struct {
	Id         string       `json:"id"`
	Name       string       `json:"name"`
	Token      string       `json:"token,omitempty"`
	ProjectId  string       `json:"project_id"`
	Scope      string       `json:"scope"`
	CreatedAt  time.Time    `json:"created_at"`
	TokenUsage []TokenUsage `json:"usage"`
}

type TokenUsage struct {
	BillableTime int64     `json:"billable_time"`
	FirstUsed    time.Time `json:"first_used"`
	LastUsed     time.Time `json:"last_used"`
	Instance     string    `json:"instance"`
	Jobs         int64     `json:"jobs"`
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

type FfmpegTemplate struct {
	// video common x264, x265 and av1
	Width        int64   `json:"width,omitempty"`
	Height       int64   `json:"height,omitempty"`
	Framerate    float64 `json:"framerate,omitempty"`
	Gop          int64   `json:"gop,omitempty"`
	Duration     int64   `json:"duration,omitempty"`
	VideoBitrate int64   `json:"video_bitrate,omitempty"`
	DisableVideo bool    `json:"disable_video,omitempty"`
	AudioBitrate int64   `json:"audio_bitrate,omitempty"`
	Channels     int64   `json:"channels,omitempty"`
	DisableAudio bool    `json:"disable_audio,omitempty"`
	Minrate      int64   `json:"minrate,omitempty"`
	Maxrate      int64   `json:"maxrate,omitempty"`
	Bufsize      int64   `json:"bufsize,omitempty"`
	Chunk        int64   `json:"chunk,omitempty"`
	PixFmt       string  `json:"pixfmt,omitempty"`
	Seek         int64   `json:"seek,omitempty"`
	Crf          int64   `json:"crf,omitempty"`
	X264KeyInt   int64   `json:"x264_keyint,omitempty"`
	X265KeyInt   int64   `json:"x265_keyint,omitempty"`
	Level        int64   `json:"level,omitempty"`
	Profilev     string  `json:"profilev,omitempty"`
	Preset       string  `json:"preset,omitempty"`

	// hls
	HlsTime        int64  `json:"hls_time,omitempty"`
	HlsSegmentType string `json:"hls_segment_type,omitempty"`
	HlsEnc         bool   `json:"hls_enc,omitempty"`
	HlsEncKey      string `json:"hls_enc_key,omitempty"`
	HlsEncKeyUrl   string `json:"hls_enc_key_url,omitempty"`
	HlsEncIv       string `json:"hls_enc_iv,omitempty"`

	// image
	Interval int64 `json:"interval,omitempty"`
	Sprite   bool  `json:"sprite,omitempty"`
	Frames   int64 `json:"frames,omitempty"`
}
