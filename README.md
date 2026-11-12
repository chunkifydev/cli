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
  <a href="https://github.com/chunkifydev/cli/actions/workflows/test.yml"><img alt="Build status" src="https://img.shields.io/github/actions/workflow/status/chunkifydev/cli/test.yml?style=flat-square&branch=main" /></a>
</p>
<p align="center">
<img src="https://github.com/user-attachments/assets/40ec208e-d416-4b08-85fb-029b0a42c191" alt="demo" width="640" />
</p>

The Chunkify CLI brings super-fast video transcoding to your terminal. With a single command, you can upload local files, transcode videos using Chunkify's parallel technology, and download the processed files to your local disk.

For local development, the Chunkify CLI provides a convenient command to [forward webhook notifications](#chunkify-api-integration) to your local application URL.

**Useful links:**

- [chunkify.dev](https://chunkify.dev)
- [ CLI documentation](https://chunkify.dev/docs/cli)
- [Documentation](https://chunkify.dev/docs)
- [Dashboard](https://chunkify.dev/~)
- [Sign up](https://chunkify.dev/signup)


## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Authentication](#authentication)
- [Quick Start with Chunkify](#quick-start-with-chunkify)
  - [Transcode a Video](#transcode-a-video)
  - [Read from connected storage](#read-from-connected-storage)
  - [Per-title encoding](#per-title-encoding)
  - [Extract audio](#extract-audio)
- [Transcoding Parameters](#transcoding-parameters)
  - [Video Settings](#video-settings)
  - [Audio Settings](#audio-settings)
  - [H.264/H.265/AV1 Settings](#h264h265av1-settings)
  - [VP9 Settings](#vp9-settings)
- [JSON Output](#json-output)
- [CLI Profiles](#cli-profiles)
- [Direct API commands](#direct-api-commands)
- [Chunkify API Integration](#chunkify-api-integration)
  - [Receiving Webhook Notifications Locally](#receiving-webhook-notifications-locally)
    
## Prerequisites

You need to have a Chunkify account to use the CLI. If you don't have one, you can sign up for a free trial at [chunkify.dev](https://chunkify.dev).

## Installation

### npm and npx

With Node.js 22 or later, run the CLI without a global install:

```bash
npx @chunkify/cli@latest --help
```

Or install the `chunkify` command globally:

```bash
npm install -g @chunkify/cli
chunkify version
```

The npm package supports macOS and Linux on x64 and ARM64, and Windows on x64. It includes the native binaries, so installation does not need Go, curl, or npm install scripts.

Use `npx @chunkify/cli@latest` in place of `chunkify` in the examples below. To pin a version, use `npx @chunkify/cli@<version>` with a version published to npm.

Update a global install with `npm install -g @chunkify/cli@latest`. For a project dependency, use `npm install --save-dev @chunkify/cli` and run it with `npx chunkify`.

### Shell installer

On macOS and Linux, install the latest version without Node.js:

```
curl -fsSL https://cli.chunkify.sh | bash
```

## Authentication

After the installation, the first step is to set your project token:

```
chunkify config token <sk_project_token>
```

> [!TIP]
> You will find the project token in your project settings page under the **Project access token** section. It's best to create a new token for the CLI.

Another way to authenticate is to set the environment variable `CHUNKIFY_TOKEN`:

```bash
export CHUNKIFY_TOKEN=sk_project_token
```

If you have multiple projects that you want to use with the CLI, simply use the `--profile` flag to use a different project token. See [CLI Profiles](#cli-profiles) for more details.

## Quick Start with Chunkify

You can use the Chunkify CLI to transcode a local video, an HTTP URL, an existing source ID, or an object in connected storage using `store://`.

### Transcode a Video

```
chunkify -i video.mp4 -o video_1080p.mp4 -f mp4_h264 -s 1920x1080 --crf 21
```

It will upload the video to Chunkify, transcode it to MP4 H.264, and download it to your local disk.

Transcoding uses the Chunkify [Transcodings API](https://chunkify.dev/docs). Each run creates one transcoding and downloads its output file.

For local files, the CLI creates an upload session, transfers the file, and calls the completion endpoint before looking up the source. Files over 100 MiB use a multipart transfer: the CLI uploads parts in parallel and retries a failed part without restarting the whole file. Temporary completion failures are retried without uploading the file again. Both requests must finish before the session expires. See the [video upload guide](https://chunkify.dev/docs/integration/video-upload).

Save a storage connection for the CLI to use for both uploads and transcoding outputs:

```bash
chunkify config storage-id stor_aws_example
chunkify -i video.mp4 -o output.mp4
```

The configured ID takes precedence over both `--upload-storage-id` and `--output-storage-id`. For external storage, the CLI generates `chunkify-cli/sources/<execution-id>/<input-filename>` for uploads and `chunkify-cli/transcodings/<execution-id>/<output-filename>` for outputs. When no local output filename is supplied, it uses `output` with the format's extension. Explicit `--upload-storage-path` and `--output-storage-path` values are preserved. Chunkify-managed storage, identified by `stor_chunkify_*`, uses API-generated paths and rejects explicit path flags for the corresponding operation.

`--storage-path` remains a deprecated alias for `--output-storage-path`. Existing commands still work and print a deprecation warning.

Without a configured storage ID, the CLI sends the IDs and paths supplied on the command line. Omitted IDs use the project's default storage. If that storage is external, you must supply `--upload-storage-path` for a local upload and `--output-storage-path` for transcoding outputs. Upload paths are exact bucket keys; output paths are relative to the storage connection's `base_prefix`.

Selecting upload storage through config or flags forces a new upload, even if the video was uploaded before. Generated paths use an execution ID to avoid collisions. An explicit path can overwrite an existing object.

> [!TIP]
> When transcoding the same local video multiple times, we use the source already created on Chunkify so you won't need to upload the video more than once.

You can also transcode a video that is publicly available via HTTP:

```
chunkify -i https://cdn/video.mp4 -o video_1080p.mp4 -f mp4_h264 -s 1920x1080 --crf 21
```

If a video has already been uploaded to Chunkify, you can simply use the source ID as the input:

```
chunkify -i src_33aoGbF6fyY49qUVebIeNaxZJ34 \
         -o video_av1_1080p.mp4 \
         -f mp4_av1 \
         -s 1920x1080 \
         --crf 34 \
         --preset 7
```

> [!TIP]
> If `--format` is omitted but `--output` is set, we will match the file extension to the appropriate format:
> 
> - `.mp4` → `mp4_h264`
> - `.webm` → `webm_vp9`
> - `.m4a` → `m4a_aac`

Sometimes, it's better to know what the input specifications are before transcoding. Use `--input` without setting `--format`, and it will only upload or make available the source video:

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

Now you can perfectly adapt your transcoding settings to your needs with a second command by either setting `--input` to the source ID or the same local file (if uploaded from disk).

### Read from connected storage

Use `store://` to read an existing object directly from external storage. This creates a source without uploading or copying the file:

```bash
chunkify -i store://videos/input.mp4 --source-storage-id stor_aws_example
```

To transcode and download the result using a saved storage connection:

```bash
chunkify config storage-id stor_aws_example
chunkify -i store://videos/input.mp4 -o output.mp4
```

For source inputs, storage selection uses `--source-storage-id` first, then config `storage-id`, then the project's default storage. An explicit source storage ID takes precedence over config because it identifies the bucket containing the existing object. Output storage continues to follow the output rules above.

Everything after `store://` is the exact object key, including any leading slash. The CLI does not add `base_prefix`, decode URL escapes, or normalize the path. Quote the input if the key contains spaces or shell characters, for example `-i 'store://videos/my clip.mp4'`. Keys must be 1 to 1024 UTF-8 bytes.

The source storage must be external; `stor_chunkify_*` cannot be used for this input mode. The object must remain available while Chunkify processes it. `--source-storage-id` is only valid with `store://`, and upload storage flags cannot be combined with `store://`.

### Per-title encoding

Use `--per-title` to let Chunkify select video quality and bitrate settings for your source and output resolution. It supports the MP4 and WebM video formats and is disabled by default.

To encode a 1080p H.264 video with per-title optimization:

```bash
chunkify -i video.mp4 \
         -o video_1080p.mp4 \
         -f mp4_h264 \
         -s 1920x1080 \
         --per-title
```

Do not combine `--per-title` with `--crf`, `--vb`, `--maxrate`, or `--bufsize`. Audio-only `m4a_aac` output does not support per-title encoding.

### Extract audio

Use `m4a_aac` to extract the audio track as AAC in an M4A file:

```bash
chunkify -i video.mp4 -o audio.m4a -f m4a_aac --ab 192K --ar 48000
```

Video settings are rejected for `m4a_aac`.

## Transcoding Parameters

| Flag | Type | Description |
|------|------|-------------|
| `-i, --input` | string | Local file, HTTP URL, source ID (`src_*`), or `store://object-key` |
| `--source-storage-id` | string | External storage for `store://` input; takes precedence over config `storage-id` |
| `-o, --output` | string | Output file path |
| `-f, --format` | string | `mp4_h264`, `mp4_h265`, `mp4_av1`, `webm_vp9`, `m4a_aac` |
| `--upload-storage-id` | string | Storage for local uploads; overridden by config `storage-id` |
| `--output-storage-id` | string | Storage for transcoding outputs; overridden by config `storage-id` |
| `--upload-storage-path` | string | Exact upload object key; generated for external storage when config `storage-id` is set |
| `--output-storage-path` | string | Output path relative to `base_prefix`; generated for external storage when config `storage-id` is set |

### Video Settings

| Flag | Type | Description | Value |
|------|------|-------------|-------|
| `-s, --resolution` | string | Set resolution wxh | 0-7680x0-7680 |
| `-r, --framerate` | float | Set frame rate | 0.001-240 |
| `-g, --gop` | int | Set group of pictures size | 1-960 |
| `--vb` | int | Set video bitrate in bits per second | 100000-50000000. You can also use units like 2000K or 2M |
| `--maxrate` | string | Set maximum bitrate in bits per second | 100000-50000000. You can also use units like 2000K or 2M |
| `--bufsize` | string | Set buffer size in bits | 100000-50000000. You can also use units like 2000K or 2M |
| `--pixfmt` | string | Set pixel format | H.264: yuv420p, yuv422p, yuv444p, yuv420p10le, yuv422p10le, yuv444p10le. H.265: H.264 values plus yuv420p12le, yuv422p12le, yuv444p12le. AV1: yuv420p, yuv420p10le. VP9: yuv420p, yuv422p, yuv440p, yuv444p and their 10le/12le variants |
| `-t, --duration` | float | Set output duration in seconds |
| `--seek` | float | Start at this position in seconds |
| `--vn` | bool | Disable video |
| `--per-title` | bool | Automatically select video rate-control settings for each source | Disabled by default |

See [Per-title encoding](#per-title-encoding) for examples and compatible settings.

### Audio Settings

| Flag | Type | Description | Value |
|------|------|-------------|-------|
| `--ab` | int | Set audio bitrate in bits per second | 32000-512000. You can also use units like 128K |
| `--ar` | int | Set audio sample rate in Hz | 8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000, 64000, 88200, 96000 |
| `--channels` | int | Set number of audio channels | 1, 2, 6 (5.1), 8 (7.1) |
| `--an` | bool | Disable audio |

### H.264/H.265/AV1 Settings

| Flag | Type | Description | Value |
|------|------|-------------|-------|
| `--crf` | int | Set constant rate factor | H.264/H.265: 16-35, AV1: 16-63, VP9: 15-35 |
| `--preset` | string | Set encoding preset | H.264/H.265: ultrafast, superfast, veryfast, faster, fast, medium, AV1: 6-13 |
| `--profilev` | string | Set video profile | H.264: baseline, main, high, high10, high422, high444, H.265: main, main10, main12, main422-10, main422-12, main444-8, main444-10, main444-12, AV1: main |
| `--level` | int | Set encoding level | H.264: 10, 11, 12, 13, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, H.265: 30, 31, 41, AV1: 30, 31, 41 |
| `--x264keyint` | int | H.264 - Set x264 keyframe interval | 1-960 |
| `--x265keyint` | int | H.265 - Set x265 keyframe interval | 1-300 |

### VP9 Settings

| Flag | Type | Description | Value |
|------|------|-------------|-------|
| `--quality` | string | Set VP9 quality | good, best, realtime |
| `--cpu-used` | string | Set VP9 CPU usage | 0-8 |

## JSON Output

It's possible to output the progress in JSON format by passing the `--json` flag.

```
chunkify -i video.mp4 -o video_1080p.mp4 -s 1920x1080 --crf 21 --json
```

```json
{"status":"Queued","progress":0,"fps":0,"speed":"N/A","out_time":0,"eta":"N/A"}
{"status":"Encoding","progress":20,"fps":0,"speed":"N/A","out_time":0,"eta":"N/A"}
{"status":"Ready","progress":100,"fps":0,"speed":"N/A","out_time":0,"eta":"N/A"}
{"status":"Transcoding","progress":40,"fps":100,"speed":"5x","out_time":3,"eta":""}
{"status":"Transcoding","progress":70,"fps":100,"speed":"5x","out_time":3,"eta":""}
{"status":"Merging","progress":90,"fps":12,"speed":"1.2x","out_time":3,"eta":""}
{"status":"Merging","progress":100,"fps":12,"speed":"1.2x","out_time":3,"eta":""}
{"status":"Downloading","progress":100,"fps":0,"speed":"105MB/s","out_time":0,"eta":"0s"}
{"status":"Completed","progress":0,"fps":0,"speed":"","out_time":0,"eta":""}
```

## CLI Profiles

Storage configuration is saved per profile, like the project token:

```bash
chunkify config storage-id stor_aws_example --profile testing
chunkify config storage-id --profile testing       # Show the saved ID
chunkify -i video.mp4 -o output.mp4 --profile testing
chunkify config storage-id "" --profile testing    # Clear the saved ID
```

This setting selects storage for CLI requests; it does not follow later changes to the project's default storage in the dashboard. Clear it to use the storage ID flags or the API default again.

You may have multiple projects and want to use different project tokens for different tasks, or simply to differentiate between different environments.

The CLI provides a global `--profile` flag to use different project tokens.

First, let's save a new token for the `testing` profile:

```
chunkify config token sk_project_token --profile testing
```

Now you can use this profile with the `--profile` flag for transcoding:

```
chunkify -i video.mp4 -o video_1080p.mp4 -s 1920x1080 --crf 21 --profile testing
```

> [!NOTE]
> If no profile given, the CLI will use the default one

## Direct API commands

Use `chunkify api <resource> <action>` to call the Chunkify API. These commands print JSON to standard output by default, with no banner or update notice. The CLI reads the [production OpenAPI definition](https://chunkify.dev/docs/openapi.json) to discover resources and actions, and caches it for one hour. If the docs site is unavailable, the last cached definition is used. Use `chunkify api --help` to list resources and `chunkify api projects --help` to list a resource's actions.

Command names follow OpenAPI `operationId` values: `updateProject` becomes `projects update`, and `getJobFiles` becomes `job-files list` because its route returns a collection. New action verbs follow the same rule.

```bash
chunkify api jobs list --query limit=10 | jq '.data'
chunkify api jobs get job_123
chunkify api projects update my-project --data '{"name":"New name"}'
chunkify api projects update my-project --data '{"storage_id":"stor_aws_example"}'
chunkify api storages create --data @storage.json
```

Use `--data @-` to read a JSON body from standard input. Repeat `--query key=value` for multiple query parameters. Path IDs follow the action name, as shown by each command's `--help` output. The CLI prints the API's JSON response unchanged. If the API returns no body, it prints `{}`. Non-JSON responses, such as an asset preview, are printed as `{"encoding":"base64","data":"..."}`. Errors are JSON on standard error and cause a nonzero exit status.

Project resources use the project token configured with `chunkify config token` or `CHUNKIFY_TOKEN`. Projects and access tokens use a team token:

```bash
chunkify config team-token <sk_team_token>
# Or set CHUNKIFY_TEAM_TOKEN in the environment.
chunkify api projects list
```

Both tokens can be saved under the same `--profile`. Team commands require only a team token; project commands require only a project token. The CLI stores configured tokens in the system keyring and masks them when showing config values.

The request URL comes from `chunkify config endpoint` for the selected profile, or `CHUNKIFY_ENDPOINT` when set. The default is `https://api.chunkify.dev/v1`. The URL inside the OpenAPI document does not change where requests go.

To use another OpenAPI definition for command discovery, save its URL in the same profile:

```bash
chunkify config endpoint https://staging-api.example.com/v1 --profile staging
chunkify config openapi-url https://staging-api.example.com/openapi.json --profile staging
chunkify api projects list --profile staging
```

`CHUNKIFY_OPENAPI_URL` overrides the saved OpenAPI URL. Without either setting, the CLI uses the production definition. Clear a saved URL with `chunkify config openapi-url "" --profile staging`.

## Chunkify API Integration

### Receiving Webhook Notifications Locally

When integrating Chunkify into your app, you must rely on webhooks to receive events when a transcoding is ready or when an upload is created. We have added the `listen` command to forward webhooks to your local server URL, which is normally not available publicly.

> [!NOTE]
> First, you need to retrieve your webhook secret in your project settings page under the Webhooks section.

<p align="center">
<picture width="600">
  <source srcset="https://github.com/user-attachments/assets/e3617da9-ad3b-4c57-bbfc-a30f51220b12" media="(prefers-color-scheme: light)">
  <source srcset="https://github.com/user-attachments/assets/a274b4a4-89ba-4874-8dfb-748d9984145c" media="(prefers-color-scheme: dark)">
  <img width="600" alt="webhook secret" src="https://github.com/user-attachments/assets/e3617da9-ad3b-4c57-bbfc-a30f51220b12" />
</picture>
</p>

Start forwarding webhooks to your local server:

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
- all (35 events)

────────────────────────────────────────────────

[200 OK] notf_33f3pVlO3782tPF9CkioGK1IKTu transcoding.ready (trn_33f3ocg9Vg0o0gDgg5JpCCR3DzX)
[200 OK] notf_33f3tiGWDw78SLHefdswGaL7UpB transcoding.ready (trn_33f3siy9JhMrlsIY69q2gHqL3bh)
```

By default, it forwards every event the API delivers: `job.*`, `source.*`, `upload.*`, `asset.*`, `asset.output.*`, `workflow_run.*`, `transcoding.*` and `thumbnail.*`. Unknown event names are rejected. You can specify the ones you are interested in:

```
chunkify listen \
  --forward-to http://localhost:3000/webhooks/chunkify \
  --webhook-secret <secret-key> \
  --events transcoding.ready,transcoding.failed,transcoding.cancelled
```

What `chunkify listen` does under the hood:
-   Creates a temporary webhook in your project
-   Forwards all notifications to your local server
-   Signs requests with the webhook secret key
-   Cleans up the webhook when you exit

## Development

### Prerequisites

-   Go 1.x or higher

### npm packaging

Node.js 22 or later and Go 1.23 or later are required to build the npm package. There are no npm dependencies to install.

```bash
npm test
npm run test:install
npm pack
```

`npm pack` builds all five platform binaries from this checkout. `npm run test:install` packs the CLI, installs it into a temporary global prefix with install scripts disabled, and runs it through both the global command and npm exec, the command behind npx. CI runs these checks on macOS, Linux, and Windows.

Release Please updates `package.json` alongside the Go release version. The release workflow attaches `chunkify-cli-<version>.tgz` to each GitHub release. Each npm package contains the matching CLI version.

Before the first npm release, a maintainer must have publishing access to the `@chunkify` npm scope. Download the `.tgz` from the intended GitHub release, sign in with `npm login`, and publish that file with `npm publish ./chunkify-cli-<version>.tgz --access public`.

Then configure an [npm trusted publisher](https://docs.npmjs.com/trusted-publishers/) in the package settings for GitHub owner `chunkifydev`, repository `cli`, and workflow `release.yml`. Allow direct publishing with `npm publish`. Set the GitHub repository Actions variable `NPM_PUBLISH_ENABLED` to `true` to publish subsequent releases automatically. Publishing uses GitHub's identity token and does not require a stored npm token. Until enabled, the workflow only uploads the npm tarball to the GitHub release.

If npm publishing fails after a GitHub release, rerun the failed npm job. To publish an already-created tarball manually, download it from that release and use the same `npm publish` command above.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT

----------------------

<p align="center">
  <picture width="300">
    <source srcset="https://github.com/user-attachments/assets/b897dd77-7ce1-4e9b-b0b3-393ca9aa5473" media="(prefers-color-scheme: dark)">
    <source srcset="https://github.com/user-attachments/assets/d0d31788-9b13-4a75-ab5a-97831791d910" media="(prefers-color-scheme: light)">
    <img width="300" alt="chunkify-ascii-black" src="https://github.com/user-attachments/assets/b897dd77-7ce1-4e9b-b0b3-393ca9aa5473" />
  </picture>
</p>
