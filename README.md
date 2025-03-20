# Chunkify CLI

The Chunkify CLI provides easy access to all Chunkify services and features directly from your terminal.

#### With this CLI, you can:

-   **Create and manage** sources, projects, jobs, storages, webhooks, notifications and tokens
-   **View** jobs logs

## Installation

```bash
curl -L https://chunkify.dev/install.sh | sh
```

This will automatically download and install the latest version of the CLI.

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
╭──────────────────────────────────────────────────────────────────────────────────╮
│ Date                 Project Id     Name       Storage             Active        │
├──────────────────────────────────────────────────────────────────────────────────┤
│ 11 Jul 24 13:00 UTC  ch..54        Local Dev  chunkify-us-east-1   yes           │
╰──────────────────────────────────────────────────────────────────────────────────╯
```

### Authentication Commands

-   If you encounter "Unauthorized" errors, try logging in again with `chunkify auth login`
-   To logout: `chunkify auth logout`

### Environment-based Authentication

For CI/CD environments or automated scripts where interactive login isn't possible (like GitHub Actions), you can authenticate using environment variables. You'll need both an account token and a project token for full CLI functionality:

```bash
# First, create an account token
$ chunkify tokens create --scope account

# Create a project token using your project ID
CHUNKIFY_ACCOUNT_TOKEN=sk_account_123...abc chunkify tokens create --scope project --project-id your_project_id

# Use these tokens inline for any CLI command
CHUNKIFY_ACCOUNT_TOKEN=sk_account_123...abc CHUNKIFY_PROJECT_TOKEN=sk_project_456...xyz chunkify jobs list
```

> Note: Store your tokens securely and never commit them to version control. For GitHub Actions, use repository secrets. When you're done with tokens, revoke them explicitly with `chunkify tokens revoke <token-id>`.

## Usage

Here are some of the most used commands

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
$ chunkify jobs create --source-id 55..79
╭────────────────────────────────────────────────────────────────────────────────────────────────╮
│ Date                 Id      Status  Progress  Template     Transcoders  Speed  Time  Billable │
├────────────────────────────────────────────────────────────────────────────────────────────────┤
│ 20 Mar 25 13:58 UTC  31..3a  queued  0%       mp4/x264-v1   1 x 4vCPU    0.00x  00:00    -     │
╰────────────────────────────────────────────────────────────────────────────────────────────────╯
```

#### Create a job directly with a source URL:

```bash
$ chunkify jobs create --url https://videosource.com/video.mp4
```

### Webhook Creation

The webhook will be created for your current selected project.

```bash
$ chunkify webhooks create --url http://www.webhooks.com/
╭────────────────────────────────────────────────────────────╮
│ Id      Url                     Events  Active             │
├────────────────────────────────────────────────────────────┤
│ 66..3b  http://www.webhooks.com  *      yes                │
╰────────────────────────────────────────────────────────────╯
```

### Local Webhook Development

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
-   `--debug`: Enable debug mode for verbose output
-   `--endpoint`: Specify a custom API endpoint
-   `--env-project-id`: Set the project context

## Development

### Prerequisites

-   Go 1.x or higher
-   Git

## Support

For support, please visit [Chunkify Support](https://chunkify.dev/support) or open an issue in the GitHub repository.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Add your license information here]
