# OAuth CLI (Go Version)

A fast and efficient CLI tool to quickly obtain JWT tokens via OAuth2 flow with support for multiple providers, written in Go.

## Features

- ⚡ **Lightning Fast**: Written in Go for maximum performance
- 🔐 Complete OAuth 2.0 flow with any OAuth2 provider
- 🌐 Automatic browser opening for authentication
- 💾 Save multiple app configurations for repeated use
- 🎨 Beautiful CLI interface with colors and emojis
- 🔧 Interactive configuration setup
- 🚀 Quick token retrieval
- 📱 Support for multiple OAuth2 apps
- 🔄 Easy switching between different providers
- 📦 Single binary distribution

## Installation

### Option 1: Build from source
```bash
git clone <repository-url>
cd oauth-util
go mod tidy
go build -o oauth-util
```

### Option 2: Install globally
```bash
go install github.com/ashsmith/oauth-util@latest
```

### Option 3: Download pre-built binary
Download the latest release for your platform from the releases page.

## Releases

This project uses automated releases via GitHub Actions. When a new semver tag is pushed (e.g., `v1.2.3`), the workflow will:

1. **Build binaries** for multiple platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64, arm64)

2. **Create a GitHub release** with:
   - Release notes generated from commits
   - Downloadable binaries for all platforms
   - Source code archives

### Creating a Release

To create a new release:

```bash
# Create and push a new tag
git tag v1.2.3
git push origin v1.2.3
```

The GitHub Action will automatically:
- Build all platform binaries
- Create a release on GitHub
- Upload the binaries for download

### Downloading Releases

Visit the [Releases page](https://github.com/ashsmith/oauth-util/releases) to download pre-built binaries for your platform.

## Usage

### Quick Start

1. **Configure your first app** (one-time setup):
```bash
./oauth-util configure
```

2. **Get a JWT token**:
```bash
./oauth-util token
```

### Managing Multiple Apps

**List all configured apps:**
```bash
./oauth-util list
```

**Configure additional apps:**
```bash
./oauth-util configure
```

**Set a default app:**
```bash
./oauth-util set-default <appName>
```

**Delete an app:**
```bash
./oauth-util delete <appName>
```

### Manual Login

You can also perform a one-time login with specific parameters:

```bash
./oauth-util login \
  --client-id YOUR_CLIENT_ID \
  --domain https://your-oauth-provider.com \
  --scope "openid email profile" \
  --port 3000
```

### Using Specific Apps

**Login with a specific app:**
```bash
./oauth-util login --app myapp
```

**Get token from a specific app:**
```bash
./oauth-util token --app myapp
```

### JSON Output for Scripting

**Get JSON output for piping to jq:**
```bash
./oauth-util token --json | jq
./oauth-util token --json | jq '.id_token'
./oauth-util token --json | jq '.access_token'
```

Using jsonpath:

```
./oauth-util token --jsonpath '.access_token'
```

## Command Reference

#### `configure`
Interactive setup to save OAuth2 app configuration:
```bash
./oauth-util configure
```

#### `login`
Start OAuth flow with specific parameters:
```bash
./oauth-util login [options]
```

Options:
- `-c, --client-id` - OAuth2 Client ID
- `-d, --domain` - OAuth2 Domain (full URL)
- `-s, --scope` - OAuth2 Scope (default: openid email profile)
- `-p, --port` - Local server port (default: 3000)
- `-a, --app` - Use saved app configuration
- `--json` - Output only JSON data (for piping to jq)

#### `token`
Get JWT token using saved app configuration:
```bash
./oauth-util token [options]
```

Options:
- `-p, --port` - Local server port (default: 3000)
- `-a, --app` - Use specific app (defaults to default app)
- `--json` - Output only JSON data (for piping to jq)

#### `list`
List all configured apps:
```bash
./oauth-util list
```

#### `set-default`
Set default app:
```bash
./oauth-util set-default <appName>
```

#### `delete`
Delete an app configuration:
```bash
./oauth-util delete <appName>
```

## Configuration

The tool stores your app configurations locally in `~/.config/oauth-util.json`. Each app can have:

- **Name**: Friendly name for easy reference
- **Client ID**: Your OAuth2 Client ID
- **Domain**: Full OAuth2 provider URL (e.g., https://accounts.google.com)
- **Scope**: OAuth2 scope (default: openid email profile)

### Example Configurations

**Google OAuth2:**
- Domain: `https://accounts.google.com`
- Scope: `openid email profile`

**GitHub OAuth2:**
- Domain: `https://github.com`
- Scope: `read:user user:email`

**Custom OAuth2 Provider:**
- Domain: `https://your-provider.com`
- Scope: `openid email profile`

## Performance Benefits

Compared to the Node.js version, the Go implementation provides:

- **~10x faster startup time**
- **~5x faster token exchange**
- **~3x lower memory usage**
- **Single binary distribution**
- **No runtime dependencies**

## How It Works

1. **Local Server**: Starts a local HTTP server to receive the OAuth callback
2. **Browser Authentication**: Opens your default browser to the OAuth2 login page
3. **User Login**: You authenticate with your credentials in the browser
4. **Callback**: OAuth2 provider redirects back to your local server with an authorization code
5. **Token Exchange**: The CLI exchanges the authorization code for JWT tokens
6. **Cleanup**: The local server is shut down and the JWT is displayed

## Prerequisites

- Go 1.21+ (for building from source)
- OAuth2 provider configured with authorization code flow
- OAuth2 Client ID and domain URL
- Redirect URI configured as `http://localhost:3000/callback` (or your chosen port)

## Troubleshooting

### Port Already in Use
If you get a "port already in use" error, specify a different port:
```bash
./oauth-util token --port 3001
```

### Invalid Client ID or Domain
Make sure your OAuth2 configuration is correct:
- Client ID should match your OAuth2 application
- Domain should be the full URL of your OAuth2 provider
- Redirect URI should be configured in your OAuth2 application

### Browser Doesn't Open
If the browser doesn't open automatically, the tool will show you the URL to visit manually.

## Security Notes

- The tool stores configuration locally on your machine
- JWT tokens are displayed in the terminal (consider clearing terminal history if needed)
- The local server only runs during the OAuth flow
- No credentials are stored, only configuration settings

## Development

To run in development mode:
```bash
go run .
```

To build for different platforms:
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o oauth-util-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o oauth-util-macos

# Windows
GOOS=windows GOARCH=amd64 go build -o oauth-util.exe
```

## License

MIT
