<p align="center">
  <a href="https://chunkify.dev">
    <picture width="300">
      <source srcset="https://github.com/user-attachments/assets/63a8660c-71be-4ded-9c2d-000097195680" media="(prefers-color-scheme: dark)">
      <source srcset="https://github.com/user-attachments/assets/3f9ad6f4-43e3-483e-943a-d20a7873c2a6" media="(prefers-color-scheme: light)">
      <img width="300" alt="chunkify-black" src="https://github.com/user-attachments/assets/2cece349-77b6-4a13-badf-512ae11ca5e2" />
    </picture>
  </a>
</p>
<p align="center">The next generation Cloud transcoding service.</p>

<p align="center">
  <a href="https://github.com/chunkifydev/cli/releases"><img src="https://img.shields.io/github/release/chunkifydev/cli.svg" alt="Latest Release"></a>
  <a href="https://github.com/chunkifydev/cli/actions/workflows/release.yml"><img alt="Build status" src="https://img.shields.io/github/actions/workflow/status/chunkifydev/cli/release.yml?style=flat-square&branch=main" /></a>
</p>
<p align="center">
<img src="https://github.com/user-attachments/assets/28441a45-cf41-4476-8920-5b641812d56a" alt="demo" width="640" />
</p>

The Chunkify CLI brings super-fast video transcoding to your terminal. With a single command, you can upload local files, transcode videos using Chunkify's parallel technology, and download the processed files to your local disk.

For local development, the Chunkify CLI provides a convenient command to [forward webhook notifications](#chunkify-api-integration) to your local application URL.

Useful links:

- [chunkify.dev](https://chunkify.dev)
- [Documentation](https://chunkify.dev/docs)
- [Dashboard](https://chunkify.dev/~)
- [Sign up](https://chunkify.dev/signup)


## Table of Contents

- [Installation](#installation)
- [Authentication](#authentication)
- [Quick Start with Chunkify](#quick-start-with-chunkify)
  - [Transcode a Video](#transcode-a-video)
  - [HLS Packaging](#hls-packaging)
  - [Generate Thumbnails](#generate-thumbnails)
- [Transcoding Parameters](#transcoding-parameters)
  - [Video Settings](#video-settings)
  - [Audio Settings](#audio-settings)
  - [H.264/H.265/AV1 Settings](#h264h265av1-settings)
  - [VP9 Settings](#vp9-settings)
  - [HLS Settings](#hls-settings)
  - [JPG Settings](#jpg-settings)
- [JSON Output](#json-output)
- [Chunkify API Integration](#chunkify-api-integration)
  - [Receiving Webhook Notifications Locally](#receiving-webhook-notifications-locally)
    
## Installation

Installing the latest version:

```
curl -L https://chunkify.dev/install.sh | sh
```

## Authentication

1. After the installation, the first step is to authenticate with your Chunkify account:

```
chunkify auth login
```

2. The CLI will open your browser and ask you to select one of your teams for authentication.

3. After authentication, the CLI will prompt you to select your project.

Another way to authenticate is to setup environment variables with your project token:

```bash
export CHUNKIFY_PROJECT_TOKEN=sk_project_token
```

## Quick Start with Chunkify

You can use the chunkify CLI to transcode a local video, a URL, or a source ID if it was already uploaded to Chunkify.

### Transcode a Video

```
chunkify -i video.mp4 -o video_1080p.mp4 -f mp4/h264 -s 1920x1080 --crf 21
```

It will upload the video to Chunkify, transcode it into MP4 H264 and download it to your local disk.

By default, the number of transcoders and their type will be selected automatically according to the input and output specs.
To define them yourself use `--transcoders` and `--vcpu` like this:

```
chunkify -i video.mp4 \
         -o video_720p.mp4 \
         -f mp4/h264 \
         -s 1280x720 \
         --crf 24 \
         --transcoders 10 \
         --vcpu 8
```

> [!TIP]
> When transcoding the same local video multiple times, we use the source already created on Chunkify so you won't upload the video more than once.

You can also transcode a video publicly available via HTTP:

```
chunkify -i https://cdn/video.mp4 -o video_1080p.mp4 -f mp4/h264 -s 1920x1080 --crf 21
```

If a video has already been uploaded to Chunkify, you can simply use the source ID as input:

```
chunkify -i src_33aoGbF6fyY49qUVebIeNaxZJ34 \
         -o video_av1_1080p.mp4 \
         -f mp4/av1 \
         -s 1920x1080 \
         --crf 34 \
         --preset 7
```

> [!TIP]
> If `--format` is omitted, but the `--output` is set, we will match the file extension to the appropriate format:
> 
> - `.mp4` → `mp4/h264`
> - `.webm` → `webm/vp9`
> - `.m3u8` → `hls/h264`
> - `.jpg` → `jpg`

Sometimes, it's better to know what the input specs are before transcoding. Use `--input` without setting `--format` and it will only upload / make available the source video:

```
chunkify -i chunkify-animation-logo.mp4

  ██   ▗▄▄▖▗▖ ▗▖▗▖ ▗▖▗▖  ▗▖▗▖ ▗▖▗▄▄▄▖▗▄▄▄▖▗▖  ▗▖
██    ▐▌   ▐▌▄▐▌▐▌ ▐▌▐▛▚▖▐▌▐▌▗▞▘  █  ▐▌▗▖  ▝▚▞▘
  ██  ▝▚▄▄▖▐▌ ▐▌▝▚▄▞▘▐▌  ▐▌▐▌ ▐▌▗▄█▄▖▐▌     ▐▌

Chunkify CLI version: dev
https://chunkify.dev

────────────────────────────────────────────────

▮ Source: chunkify-animation-logo.mp4
  Duration: 00:03 Size: 61KB Video: h264, 400x400, 149KB/s, 24.00fps

────────────────────────────────────────────────

Source ID: src_33dLly8jh7bQxVJ5L9LeMG3FAVc
```

Now you can perfectly adapt your transcoding settings to your needs with a second command by either setting the `--input` to the source ID or the same local file (if uploaded from disk).

### HLS Packaging

Chunkify supports 3 HLS formats: `hls/h264` `hls/h265` and `hls/av1`.

> [!WARNING]
> Keyframes must be aligned for all renditions, so you must use the same values for `--gop`, `--x264keyint` (H264), `--x265keyint` (H265). For `hls/av1`, only `--gop` is necessary.

```
chunkify -i video.mp4 \
         -o video_540p.m3u8 \
         -f hls/h264 \
         -s 540x0 \
         -g 120 \
         --x264keyint 120 \
         --vb 800k \
         --ab 128k
```

Once the video is transcoded, the CLI will return a summary including the `HLS Manifest ID` which we will use for the next command:

```
chunkify -i video.mp4 \
         -o video_720p.m3u8 \
         -f hls/h264 \
         -s 720x0 \
         -g 120 \
         --x264keyint 120 \
         --vb 1200k \
         --ab 128k \
         --hls-manifest-id hls_33atK0NkjF3lz6qUNi3GLwYdi0m
```

> [!NOTE]
> The video bitrate and/or the audio bitrate are mandatory for HLS output

Now we have 2 renditions that belong to the same manifest:

```
manifest.m3u8
video_540p.mp4
video_540p.m3u8
video_720p.mp4
video_720p.m3u8
```

### Generate Thumbnails

To generate thumbnails every 10 seconds:

```
chunkify -i video.mp4 -o thumbnails.jpg -f jpg -s 320x0 --interval 10
```

If many thumbnails are required, it's recommended to generate a sprite image instead of multiple individual images. A sprite image is a single image containing many thumbnails arranged in a grid, which is more efficient when there are hundreds of them to download for displaying a preview.

```
chunkify -i video.mp4 -o sprite.jpg -f jpg -s 160x0 --interval 5 --sprite
```

> [!NOTE]
> For all JPG outputs, an `images.vtt` is generated which can be loaded by an HTML5 player to display a mini preview when hovering the player progress bar

The VTT filename is always `images.vtt`. Here is how it looks like:

```
WEBVTT


00:00:00.000 --> 00:00:05.000
sprite-00000.jpg#xywh=0,0,160,160

00:00:05.000 --> 00:00:10.000
sprite-00000.jpg#xywh=160,0,160,160

00:00:10.000 --> 00:00:15.000
sprite-00000.jpg#xywh=320,0,160,160
```

## Transcoding Parameters

| Flag | Type | Description |
|------|------|-------------|
| `-i, --input` | string | Input video to transcode. It can be a file, HTTP URL or source ID (src_*) |
| `-o, --output` | string | Output file path | - |
| `-f, --format` | string | `mp4/h264`, `mp4/h265`, `mp4/av1`, `webm/vp9`, `hls/h264`, `hls/h265`, `hls/av1`, `jpg` |
| `--transcoders` | int | Number of transcoders to use | 
| `--vcpu` | int | vCPU per transcoder (4, 8, or 16) |

### Video Settings

| Flag | Type | Description | Value |
|------|------|-------------|-------|
| `-s, --resolution` | string | Set resolution wxh | 0-8192x0-8192 |
| `-r, --framerate` | float | Set frame rate | 15-120 |
| `-g, --gop` | int | Set group of pictures size | 1-300 |
| `--vb` | int | Set video bitrate in bits per second | 100000-50000000. You can also use units like 2000K or 2M |
| `--maxrate` | string | Set maximum bitrate in bits per second | 100000-50000000. You can also use units like 2000K or 2M |
| `--bufsize` | string | Set buffer size in bits | 100000-50000000. You can also use units like 2000K or 2M |
| `--pixfmt` | string | Set pixel format | yuv410p, yuv411p, yuv420p, yuv422p, yuv440p, yuv444p, yuvJ411p, yuvJ420p, yuvJ422p, yuvJ440p, yuvJ444p, yuv420p10le, yuv422p10le, yuv440p10le, yuv444p10le, yuv420p12le, yuv422p12le, yuv440p12le, yuv444p12le, yuv420p10be, yuv422p10be, yuv440p10be, yuv444p10be, yuv420p12be, yuv422p12be, yuv444p12be |
| `--vn` | bool | Disable video | - |

### Audio Settings

| Flag | Type | Description | Value |
|------|------|-------------|-------|
| `--ab` | int | Set audio bitrate in bits per second | 32000-512000. You can also use units like 128K |
| `--channels` | int | Set number of audio channels | 1, 2, 5, 7 |
| `--an` | bool | Disable audio | - |

### H.264/H.265/AV1 Settings

| Flag | Type | Description | Value |
|------|------|-------------|-------|
| `--crf` | int | Set constant rate factor | H264/H265: 16-35, AV1: 16-63, VP9: 15-35 |
| `--preset` | string | Set encoding preset | H264/H265: ultrafast, superfast, veryfast, faster, fast, medium, AV1: 6-13 |
| `--profilev` | string | Set video profile | H264: baseline, main, high, high10, high422, high444, H265/AV1: main, main10, mainstillpicture |
| `--level` | int | Set encoding level | H264: 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, H265: 30, 31, 41, AV1: 30, 31, 41 |
| `--x264keyint` | int | H264 - Set x264 keyframe interval | - |
| `--x265keyint` | int | H265 - Set x265 keyframe interval | - |

### VP9 Settings

| Flag | Type | Description | Value |
|------|------|-------------|-------|
| `--quality` | string | Set VP9 quality | good, best, realtime |
| `--cpu-used` | string | Set VP9 CPU usage | 0-8 |

### HLS Settings

| Flag | Type | Description | Value |
|------|------|-------------|-------|
| `--hls-manifest-id` | string | Set HLS manifest ID | - |
| `--hls-time` | int | Set HLS segment duration in seconds | 1-10 |
| `--hls-segment-type` | string | Set HLS segment type | mpegts, fmp4 |
| `--hls-enc` | bool | Enable HLS encryption | - |
| `--hls-enc-key` | string | Set HLS encryption key | - |
| `--hls-enc-key-url` | string | Set HLS encryption key URL | - |
| `--hls-enc-iv` | string | Set HLS encryption IV | - |

### JPG Settings

| Flag | Type | Description | Value |
|------|------|-------------|-------|
| `--interval` | int | Set frame extraction interval in seconds | 1-60 |
| `--sprite` | bool | Generate sprite sheet instead of multiple JPG files | - |

## JSON Output

It's possible to output the progress in JSON format by passing the `--json` flag.

```
chunkify -i video.mp4 -o video_1080p.mp4 -s 1920x1080 --crf 21 --json
```

```json
{"status":"Queued","progress":0,"fps":0,"speed":"0.0x","out_time":0,"eta":""}
{"status":"Queued","progress":0,"fps":0,"speed":"0.0x","out_time":3,"eta":""}
{"status":"Ingesting","progress":20,"fps":0,"speed":"0.0x","out_time":3,"eta":""}
{"status":"Transcoding","progress":40,"fps":100,"speed":"5x","out_time":3,"eta":""}
{"status":"Transcoding","progress":70,"fps":100,"speed":"5x","out_time":3,"eta":""}
{"status":"Merging","progress":90,"fps":12,"speed":"1.2x","out_time":3,"eta":""}
{"status":"Merging","progress":100,"fps":12,"speed":"1.2x","out_time":3,"eta":""}
{"status":"Downloading","progress":100,"fps":0,"speed":"105MB/s","out_time":0,"eta":"0s"}
{"status":"Completed","progress":0,"fps":0,"speed":"","out_time":0,"eta":""}
```

## Chunkify API Integration

### Receiving Webhook Notifications Locally

When integrating Chunkify into your app, you must rely on webhooks to receive events when a job is completed or when an upload is created. We have added the command `listen` to forward webhooks to your local server URL which is normally not available publicly.

> [!NOTE]
> First, you need to retrieve your webhook secret in your project settings page under Webhooks section.

<p align="center">
<picture width="600">
  <source srcset="https://github.com/user-attachments/assets/e3617da9-ad3b-4c57-bbfc-a30f51220b12" media="(prefers-color-scheme: light)">
  <source srcset="https://github.com/user-attachments/assets/a274b4a4-89ba-4874-8dfb-748d9984145c" media="(prefers-color-scheme: dark)">
  <img width="600" alt="webhook secret" src="https://github.com/user-attachments/assets/e3617da9-ad3b-4c57-bbfc-a30f51220b12" />
</picture>
</p>

Start forwarding webhooks to your local server

```
chunkify listen \
  --forward-to http://localhost:3000/webhooks/chunkify \
  --webhook-secret <secret-key>

  ██   ▗▄▄▖▗▖ ▗▖▗▖ ▗▖▗▖  ▗▖▗▖ ▗▖▗▄▄▄▖▗▄▄▄▖▗▖  ▗▖
██    ▐▌   ▐▌▄▐▌▐▌ ▐▌▐▛▚▖▐▌▐▌▗▞▘  █  ▐▌▗▖  ▝▚▞▘
  ██  ▝▚▄▄▖▐▌ ▐▌▝▚▄▞▘▐▌  ▐▌▐▌ ▐▌▗▄█▄▖▐▌     ▐▌

Chunkify CLI version: dev
https://chunkify.dev

────────────────────────────────────────────────

[mac.home] Start forwarding to http://localhost:3000/webhooks/chunkify

Events:
- job.completed
- job.failed
- job.cancelled
- upload.completed
- upload.failed
- upload.expired

────────────────────────────────────────────────

[200 OK] notf_33f3pVlO3782tPF9CkioGK1IKTu job.completed (job_33f3ocg9Vg0o0gDgg5JpCCR3DzX)
[200 OK] notf_33f3tiGWDw78SLHefdswGaL7UpB job.completed (job_33f3siy9JhMrlsIY69q2gHqL3bh)
```

By default, it will forward all events, but you can specify the ones you are interested in:

```
chunkify listen \
  --forward-to http://localhost:3000/webhooks/chunkify \
  --webhook-secret <secret-key> \
  --events job.completed,job.failed,job.cancelled
```

What `chunkify listen` does under the hood:
-   Create a temporary webhook in your project
-   Forward all notifications to your local server
-   Sign requests with the webhook secret key
-   Clean up the webhook when you exit

## Development

### Prerequisites

-   Go 1.x or higher

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT

----------------------

<p align="center">
  <picture width="300">
    <source srcset="https://github.com/user-attachments/assets/8978c909-de3f-4bca-85c2-4c5d1f595b91" media="(prefers-color-scheme: dark)">
    <source srcset="https://github.com/user-attachments/assets/939c830c-3671-496f-9d2d-3b3eedb489ba" media="(prefers-color-scheme: light)">
    <img width="300" alt="chunkify-ascii-black" src="https://github.com/user-attachments/assets/8978c909-de3f-4bca-85c2-4c5d1f595b91" />
  </picture>
</p>
