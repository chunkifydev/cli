<p align="center">
  <img src="https://chunkify.s3.us-east-1.amazonaws.com/logos/chunkify.png" alt="Chunkify Logo" width="300"/>
</p>

# Chunkify CLI

The Chunkify CLI provides easy access to all Chunkify services and features directly from your terminal.

#### With this CLI, you can:

-   **Create and manage** sources, projects, jobs, storages, webhooks, notifications and tokens
-   **Forward** webhooks notifications to your local environment
-   **View** jobs logs

## Installation

Installing the latest version:

```bash
curl -L https://chunkify.dev/install.sh | sh
```

## Getting Started

1. After the installation, the first step is to authenticate with your Chunkify account:

```bash
chunkify auth login
```

2. The CLI will open your browser and ask you to select one of your team for the authentication.

3. After authentication, the CLI will prompt you to select your project.

4. Once done, you're now ready to use all CLI features! Try some basic commands to verify your setup:

```bash
# List all the projects of your team
$ chunkify projects list
╭───────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ Date                 Project Id                        Name              Storage                              │
├───────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ 11 Jul 24 13:00 UTC  proj_A1cce6120E56e7Tu9ioP09Nhjk1  Chunkify project  stor_aws_2vuLlCyiW1mvAsKGq5zECh1MvRm │
╰───────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Authentication Commands

-   If you encounter "Unauthorized" errors, try logging in again with `chunkify auth login`
-   To logout: `chunkify auth logout`

### Environment-based Authentication

For CI/CD environments or automated scripts where interactive login isn't possible (like GitHub Actions), you can authenticate using environment variables. First, generate your tokens from the chunkify app, you'll need both an account token and a project token for full CLI functionality :

-   Team token: Required for projects and tokens commands
-   Project token: Required for all other commands (jobs, sources, webhooks, etc.)

```bash
# Option 1: Use tokens inline for a single command
$ CHUNKIFY_TEAM_TOKEN=sk_team_token chunkify projects list
$ CHUNKIFY_PROJECT_TOKEN=sk_project_token chunkify jobs list

# Option 2: Export tokens for multiple commands in the same session
$ export CHUNKIFY_TEAM_TOKEN=sk_team_token
$ export CHUNKIFY_PROJECT_TOKEN=sk_project_token
$ chunkify projects list  # Uses team token
$ chunkify jobs list      # Uses project token
$ chunkify sources list   # Uses project token
```

> Note: Store your tokens securely and never commit them to version control. For GitHub Actions, use repository secrets. When you're done with tokens, revoke them explicitly with `chunkify tokens revoke <token-id>`.

## Usage

Here are the most common commands and their usage examples:

### Source Creation

```bash
$ chunkify sources create --url https://videosource.com/video.mp4
╭────────────────────────────────────────────────────────────────────────────────────────────────╮
│ Date                 Id      Duration  Size  WxH      Video  Bitrate  Audio  Bitrate  Jobs     │
├────────────────────────────────────────────────────────────────────────────────────────────────┤
│ 20 Mar 25 13:55 UTC  97..07  00:15    2MB   1280x720 h264   1MB/s    aac    187KB/s   0        │
╰────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Job Creation

There are two ways to create a job, you can use the source id if you already have created a source or pass the source url directly.

#### Create a job using a source ID:

```bash
$ chunkify jobs create mp4/x264 --source-id src_... --height 1080 --crf 23
╭────────────────────────────────────────────────────────────────────────────────────────────────╮
│ Date                 Id      Status  Progress  Format     Transcoders  Speed  Time  Billable   │
├────────────────────────────────────────────────────────────────────────────────────────────────┤
│ 20 Mar 25 13:58 UTC  job_...  queued  0%       mp4/x264   1 x 4vCPU    0.00x  00:00    -       │
╰────────────────────────────────────────────────────────────────────────────────────────────────╯
```

#### Create a job directly with a source URL:

```bash
$ chunkify jobs create mp4/av1--url https://videosource.com/video.mp4 --width 3840 --crf 28
```

### Webhook Creation

The webhook will be created for your current selected project.

```bash
$ chunkify webhooks create --url http://www.example.com/webhooks
╭────────────────────────────────────────────────────────────╮
│ Id      Url                     Events  Active             │
├────────────────────────────────────────────────────────────┤
│ wh_...  http://www.webhooks.com  *      yes                │
╰────────────────────────────────────────────────────────────╯
```

### Testing Webhook Notifications Locally

When developing webhook integrations locally, you can use the notifications proxy to forward webhooks to your local server:

```bash
$ chunkify notifications proxy http://localhost:8000
Setting up localdev webhook...
Secret key: sk_123...abc

Start proxying notifications matching '*' to http://localhost:8000
```

The proxy will:

-   Create a temporary webhook in your project
-   Forward all notifications to your local server
-   Sign requests with the displayed secret key
-   Clean up the webhook when you exit

### Getting Help

```bash
# Get general help
$ chunkify --help

# Get help for a specific command
$ chunkify [command] --help
```

## Configuration

The CLI can be configured using command-line flags or environment variables:

### Global Flags

-   `--json`: Output results in JSON format

## Development

### Prerequisites

-   Go 1.x or higher

## Support

For support, please visit [Chunkify Support](https://chunkify.dev/support) or open an issue in the GitHub repository.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
