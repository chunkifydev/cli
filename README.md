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
<img src="https://github.com/user-attachments/assets/28441a45-cf41-4476-8920-5b641812d56a" alt="demo" width="640" />
</p>

# Chunkify CLI

The Chunkify CLI brings super-fast video transcoding to your terminal. With a single command, you can upload local files, transcode videos using Chunkify's parallel technology, and download the processed files to your local disk.

For local development, the Chunkify CLI provides a convenient command to forward webhook notifications to your local application URL.

## Installation

Installing the latest version:

```bash
$ curl -L https://chunkify.dev/install.sh | sh
```

## Getting Started

1. After the installation, the first step is to authenticate with your Chunkify account:

```bash
$ chunkify auth login
```

2. The CLI will open your browser and ask you to select one of your team for the authentication.

3. After authentication, the CLI will prompt you to select your project.

Another way to authenticate is to setup environment variables:

```bash
$ export CHUNKIFY_PROJECT_TOKEN=sk_project_token
```

## Usage

You can use the chunkify CLI to transcode a local video, an URL or a sourceID if it was already uploaded to Chunkify.

### Basic

```bash
$ chunkify -i video.mp4 -o video_1080p.mp4 -f mp4/h264 -s 1920x1080 --crf 21
```

It will upload the video to Chunkify, transcode it into MP4 H264 and download it to your local disk.

By default the number of transcoders and their type will be selected automatically according to the input and output specs.
To define them yourself use `--transcoders` and `--vcpu` like this:

```bash
$ chunkify -i video.mp4 -o video_720p.mp4 -f mp4/h264 -s 1280x720 --crf 24 --transcoders 10 --vcpu 8
```

> [!TIPS]
> When transcoding the same local video multiple times, we use the source already created on Chunkify so you won't upload the video more than once.

You can also transcode a video from an HTTP URL:

```bash
$ chunkify -i https://cdn/video.mp4 -o video_1080p.mp4 -f mp4/h264 -s 1920x1080 --crf 21
```

If a video already uploaded to Chunkify, you can simply use the source ID as input:

```bash
$ chunkify -i src_33aoGbF6fyY49qUVebIeNaxZJ34 -o video_av1_1080p.mp4 -f mp4/av1 -s 1920x1080 --crf 34 --preset 7
```

Depending on your plan, here are the available formats:

- `mp4/h264`
- `mp4/h265`
- `mp4/av1`
- `webm/vp9`
- `hls/h264`
- `hls/h265`
- `hls/av1`
- `jpg`

### HLS

To generate HLS videos, in order to perfectly align frames between multiple renditions, you must set the `gop` and `x264keyint` if H264 or `x265keyint` if H265 for *all* renditions. Here is an example:

```bash
$ chunkify -i video.mp4 -o video_540p.m3u8 -f hls/h264 -s 540x0 -g 120 --x264keyint 120 --vb 800000 --ab 128000
```

Once the video is transcoded, the CLI will return a summary including the `HLS Manifest ID` which we will use for the next command:

```bash
$ chunkify -i video.mp4 -o video_720p.m3u8 -f hls/h264 -s 720x0 -g 120 --x264keyint 120 --vb 1200000 --ab 128000 --hls-manifest-id hls_33atK0NkjF3lz6qUNi3GLwYdi0m
```

You will have the following files:

```bash
manifest.m3u8
video_540p.mp4
video_540p.m3u8
video_720p.mp4
video_720p.m3u8
```

> [!TIP]
> The video bitrate and / or the audio bitate are mandatory for HLS output

### Thumbnails

To generate thumbnails every 10 seconds:

```bash
$ chunkify -i video.mp4 -o thumbnails.jpg -f jpg -s 320x0 --interval 10
```

> [!TIP]
> If you set either `width` or `height` to 0, it will be automatically calculated according to the orignal aspect ratio.

## CLI parameters

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `-i, --input` | string | Input video to transcode. It can be a file, HTTP URL or source ID (src_*) | - |
| `-o, --output` | string | Output file path | - |
| `-f, --format` | string | Output format (mp4/h264, mp4/h265, mp4/av1, webm/vp9, hls/h264, hls/h265, hls/av1, jpg) | mp4/h264 |
| `--transcoders` | int | Number of transcoders to use | - |
| `--vcpu` | int | vCPU per transcoder (4, 8, or 16) | 8 |

### Video Settings

| Flag | Type | Description | Range |
|------|------|-------------|-------|
| `-s, --resolution` | string | Set resolution wxh | 0-8192x0-8192 |
| `-r, --framerate` | float | Set frame rate | 15-120 |
| `-g, --gop` | int | Set group of pictures size | 1-300 |
| `--vb` | int | Set video bitrate in bits per second | 100000-50000000 |
| `--maxrate` | int | Set maximum bitrate in bits per second | 100000-50000000 |
| `--bufsize` | int | Set buffer size in bits | 100000-50000000 |
| `--pixfmt` | string | Set pixel format | yuv410p, yuv411p, yuv420p, yuv422p, yuv440p, yuv444p, yuvJ411p, yuvJ420p, yuvJ422p, yuvJ440p, yuvJ444p, yuv420p10le, yuv422p10le, yuv440p10le, yuv444p10le, yuv420p12le, yuv422p12le, yuv440p12le, yuv444p12le, yuv420p10be, yuv422p10be, yuv440p10be, yuv444p10be, yuv420p12be, yuv422p12be, yuv444p12be |
| `--vn` | bool | Disable video | - |

### Audio Settings

| Flag | Type | Description | Range |
|------|------|-------------|-------|
| `--ab` | int | Set audio bitrate in bits per second | 32000-512000 |
| `--channels` | int | Set number of audio channels | 1, 2, 5, 7 |
| `--an` | bool | Disable audio | - |

### H.264/H.265/AV1 Settings

| Flag | Type | Description | Range |
|------|------|-------------|-------|
| `--crf` | int | Set constant rate factor | H264/H265: 16-35, AV1: 16-63, VP9: 15-35 |
| `--preset` | string | Set encoding preset | H264/H265: ultrafast, superfast, veryfast, faster, fast, medium, AV1: 6-13 |
| `--profilev` | string | Set video profile | H264: baseline, main, high, high10, high422, high444, H265/AV1: main, main10, mainstillpicture |
| `--level` | int | Set encoding level | H264: 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, H265: 30, 31, 41, AV1: 30, 31, 41 |
| `--x264keyint` | int | H264 - Set x264 keyframe interval | - |
| `--x265keyint` | int | H265 - Set x265 keyframe interval | - |

### VP9 Settings

| Flag | Type | Description | Range |
|------|------|-------------|-------|
| `--quality` | string | Set VP9 quality | good, best, realtime |
| `--cpu-used` | string | Set VP9 CPU usage | 0-8 |

### HLS Settings

| Flag | Type | Description | Range |
|------|------|-------------|-------|
| `--hls-manifest-id` | string | Set HLS manifest ID | - |
| `--hls-time` | int | Set HLS segment duration in seconds | 1-10 |
| `--hls-segment-type` | string | Set HLS segment type | mpegts, fmp4 |
| `--hls-enc` | bool | Enable HLS encryption | - |
| `--hls-enc-key` | string | Set HLS encryption key | - |
| `--hls-enc-key-url` | string | Set HLS encryption key URL | - |
| `--hls-enc-iv` | string | Set HLS encryption IV | - |

### JPG Settings

| Flag | Type | Description | Range |
|------|------|-------------|-------|
| `--interval` | int | Set frame extraction interval in seconds | 1-60 |
| `--sprite` | bool | Generate sprite sheet instead of multiple JPG files | - |

## Chunkify API Integration

### Receiving Webhook Notifications Locally

When integrating Chunkify into your app, you must rely on webhooks to receive events when a job is completed or when an upload is created. We have added the command `listen` to forward webhooks to your local server URL which is normally not available publicly.

```bash
$ chunkify listen --forward-to http://localhost:3000/webhooks/chunkify --webhook-secret <secret-key>
Setting up localdev webhook...

Start proxying notifications matching '*' to http://localhost:3000/webhooks/chunkify
```

> You will find the webhook secret key in your project settings under Webhooks section.

By default, it will forward all events, but you can specify the ones you are interesting to:

```bash
$ chunkify listen \
  --forward-to http://localhost:3000/webhooks/chunkify \
  --webhook-secret <secret-key> \
  --events job.completed \
  --events job.failed \
  --events job.cancelled
```

The proxy will:

-   Create a temporary webhook in your project
-   Forward all notifications to your local server
-   Sign requests with the webhook secret key
-   Clean up the webhook when you exit

### All commands and flags

```bash
$ chunkify --help

    ██   ▗▄▄▖▗▖ ▗▖▗▖ ▗▖▗▖  ▗▖▗▖ ▗▖▗▄▄▄▖▗▄▄▄▖▗▖  ▗▖
  ██    ▐▌   ▐▌▄▐▌▐▌ ▐▌▐▛▚▖▐▌▐▌▗▞▘  █  ▐▌▗▖  ▝▚▞▘
    ██  ▝▚▄▄▖▐▌ ▐▌▝▚▄▞▘▐▌  ▐▌▐▌ ▐▌▗▄█▄▖▐▌     ▐▌

  Chunkify CLI version: dev
  https://chunkify.dev

  ────────────────────────────────────────────────

Transcode videos with Chunkify

Usage:
  chunkify [flags]
  chunkify [command]

Available Commands:
  auth        Connect to your Chunkify account
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  listen      Forward webhook notifications to local HTTP URL
  version     Print the version number of Chunkify

Flags:
      --ab int                    Set audio bitrate in bits per second (32000-512000)
      --an                        Disable audio
      --bufsize int               Set buffer size in bits (100000-50000000)
      --channels int              Set number of audio channels (1, 2, 5, 7)
      --cpu-used string           Set VP9 CPU usage (0-8)
      --crf int                   Set constant rate factor (H264/H265: 16-35, AV1: 16-63, VP9: 15-35)
  -t, --duration int              Set duration in seconds
  -f, --format string             Output format (mp4/h264, mp4/h265, mp4/av1, webm/vp9, hls/h264, hls/h265, hls/av1, jpg) (default "mp4/h264")
  -r, --framerate float           Set frame rate (15-120)
  -g, --gop int                   Set group of pictures size (1-300)
  -h, --help                      help for chunkify
      --hls-enc                   Enable HLS encryption
      --hls-enc-iv string         Set HLS encryption IV
      --hls-enc-key string        Set HLS encryption key
      --hls-enc-key-url string    Set HLS encryption key URL
      --hls-manifest-id string    Set HLS manifest ID
      --hls-segment-type string   Set HLS segment type (mpegts, fmp4)
      --hls-time int              Set HLS segment duration in seconds (1-10)
  -i, --input string              Input video to transcode. It can be a file, HTTP URL or source ID (src_*)
      --interval int              Set frame extraction interval in seconds (1-60)
      --level int                 Set encoding level (H264: 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, H265: 30, 31, 41, AV1: 30, 31, 41)
      --maxrate int               Set maximum bitrate in bits per second (100000-50000000)
  -o, --output string             Output file path
      --pixfmt string             Set pixel format (yuv410p, yuv411p, yuv420p, yuv422p, yuv440p, yuv444p, yuvJ411p, yuvJ420p, yuvJ422p, yuvJ440p, yuvJ444p, yuv420p10le, yuv422p10le, yuv440p10le, yuv444p10le, yuv420p12le, yuv422p12le, yuv440p12le, yuv444p12le, yuv420p10be, yuv422p10be, yuv440p10be, yuv444p10be, yuv420p12be, yuv422p12be, yuv440p12be, yuv444p12be)
      --preset string             Set encoding preset (H264/H265: ultrafast, superfast, veryfast, faster, fast, medium, AV1: 6-13)
      --profilev string           Set video profile (H264: baseline, main, high, high10, high422, high444, H265/AV1: main, main10, mainstillpicture)
      --quality string            Set VP9 quality (good, best, realtime)
  -s, --resolution string         Set resolution wxh (0-8192x0-8192)
      --seek int                  Seek to position in seconds
      --sprite                    Generate sprite sheet
      --transcoders int           Number of transcoders to use
      --vb int                    Set video bitrate in bits per second (100000-50000000)
      --vcpu int                  vCPU per transcoder (4, 8, or 16) (default 8)
      --vn                        Disable video
      --x264keyint int            H264 - Set x264 keyframe interval
      --x265keyint int            H265 - Set x265 keyframe interval

Use "chunkify [command] --help" for more information about a command.
```

## Development

### Prerequisites

-   Go 1.x or higher

## Support

For support, please visit [Chunkify Support](https://chunkify.dev/support) or open an issue in the GitHub repository.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
