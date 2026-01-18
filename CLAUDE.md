# SMA EV Charging Log

## Project Overview

A Go CLI tool that fetches customer messages from an SMA device API and outputs specific charging events.

## Project Structure

```
├── main.go                      # Entry point
├── Dockerfile                   # Docker image definition
├── .goreleaser.yaml             # GoReleaser configuration
├── cmd/
│   ├── root.go                  # Cobra CLI setup and configuration
│   ├── sessions.go              # Sessions command (default, paired sessions)
│   ├── events.go                # Events command (raw JSON messages)
│   └── utils.go                 # Shared utilities (message filtering)
└── internal/
    ├── client/
    │   ├── client.go            # HTTP client with auto 401 retry
    │   ├── auth.go              # Bearer token fetching
    │   └── transport.go         # Logging transport middleware
    ├── log/
    │   └── log.go               # Logging configuration
    ├── models/
    │   ├── messages.go          # API request/response structs
    │   └── session.go           # Charging session struct
    └── output/
        ├── formatter.go         # Formatter interface
        ├── json.go              # JSON output formatter
        ├── csv.go               # CSV output formatter
        └── pdf.go               # PDF output formatter
```

## Libraries

- **CLI**: `github.com/spf13/cobra`
- **Configuration**: `github.com/spf13/viper` (handles flag/env var precedence)
- **Logging**: `log/slog` (standard library)
- **HTTP**: `net/http` (standard library)
- **CSV**: `encoding/csv` (standard library)
- **PDF**: `github.com/go-pdf/fpdf`

## Commands

### sessions (default)

Fetches charging events and outputs paired charging sessions. This is the default command (runs when no subcommand is specified via `rootCmd.RunE = runSessions`).

**Command-specific flags:**
- `--map-authentication` - Map authentication values (format: `old:new`, can be specified multiple times)

**Supported formats:** json, csv, pdf

### events

Fetches and outputs raw charging event messages (start/stop) without pairing them into sessions.

**Supported formats:** json only (returns error for other formats)

## Configuration

All parameters can be set via command line flags or environment variables. Command line flags take precedence.

### Global Flags (defined in root.go)

| Parameter  | Flag              | Environment Variable | Required | Notes                                              |
|------------|-------------------|----------------------|----------|---------------------------------------------------|
| Host       | `-H, --host`      | `SMA_HOST`           | Yes      | Hostname of the SMA device (defaults to https)     |
| Username   | `-u, --username`  | `SMA_USERNAME`       | Yes      | Authentication username                            |
| Password   | `-p, --password`  | `SMA_PASSWORD`       | Yes      | Authentication password                            |
| Format     | `-f, --format`    | `SMA_FORMAT`         | No       | Output format: json, csv, or pdf (default: json)   |
| Output     | `-o, --output`    | `SMA_OUTPUT`         | No       | Output file path (default: `-` for stdout)         |
| Month      | `-m, --month`     | `SMA_MONTH`          | No       | Filter by month (format: YYYY-MM)                  |
| Log Level  | `-l, --log-level` | `SMA_LOG_LEVEL`      | No       | trace, debug, info, warn, error (default: info)    |

### Sessions Command Flags (defined in sessions.go)

| Parameter          | Flag                       | Notes                                              |
|--------------------|----------------------------|----------------------------------------------------|
| Map Authentication | `-a, --map-authentication` | Map auth values (format: `old:new`, repeatable)    |

## API Endpoints

### Messages Search
- **Path**: `POST /api/v1/customermessages/search`
- **Content-Type**: `application/json`
- **Request body**:
```json
{
  "componentId": "IGULD:SELF",
  "from": null,
  "until": null,
  "messageGroupTags": [],
  "traceLevels": [],
  "marker": "<string>",
  "offset": 0
}
```
- **Response**: Array of message objects
- **Pagination**: Use `marker` from last item in response for next request; empty array indicates end

### Token
- **Path**: `POST /api/v1/token`
- **Content-Type**: `application/x-www-form-urlencoded`
- **Request body**: `grant_type=password&username=<user>&password=<pass>`
- **Response**: `{ "access_token": "<token>" }`

## Behavior

1. **URL Protocol**: If no protocol specified, defaults to `https://`
2. **Authentication**: On 401 response, automatically fetches new bearer token and retries
3. **HTTPS**: Certificate verification is disabled (`InsecureSkipVerify: true`)
4. **Filtering**: Only messages with `messageId == 9812` (charging started) or `messageId == 9813` (charging completed) are output
5. **Month Filter**: If `--month` is specified, only events with timestamps in that month are output
5. **Session Pairing**: Messages are returned newest to oldest. A charging stopped event (9813) is paired with the immediately following charging started event (9812) if present
6. **Output Formats**:
   - **JSON**: Full message object, pretty-printed (individual messages, not paired sessions)
   - **CSV/PDF**: Paired charging sessions with columns:
     - record date: Date only from charging stopped timestamp (YYYY-MM-DD format)
     - charger name: deviceName from message
     - authentication: From preceding charging started message (displayType="String", position=0), or "No Authentication"
     - start: Timestamp of charging started in ISO format (empty if no match)
     - end: Timestamp of charging stopped in ISO format
     - consumption: argument value with unitTag=8 and displayType="Fix2"
   - **PDF** additionally includes:
     - Summary: Total charging records, total consumption (kWh)
     - Landscape orientation for better column fit
7. **Trace Logging**: At `--log-level trace`, all HTTP requests/responses are logged via transport middleware

## Build & Run

```bash
# Build
go build -o sma_chg_log .

# Run sessions command (default) with JSON output
./sma_chg_log --host device.local --username admin --password secret

# Run sessions with CSV output to file
./sma_chg_log --host device.local --username admin --password secret --format csv -o sessions.csv

# Run sessions with PDF output
./sma_chg_log --host device.local --username admin --password secret --format pdf -o report.pdf

# Run sessions with month filter and authentication mapping
./sma_chg_log --host device.local --username admin --password secret --format csv --month 2026-01 \
  --map-authentication "old-id:John Doe" --map-authentication ":Unknown"

# Run events command (raw JSON messages)
./sma_chg_log events --host device.local --username admin --password secret

# Run with trace logging
./sma_chg_log --host device.local --username admin --password secret --log-level trace
```

## Development Guidelines

- **Run `go fmt ./...` after every edit to Go files**
- **Run `golangci-lint run` to check for issues**
- Keep all API-related code in `internal/client/`
- Keep data structures in `internal/models/`
- Keep output formatters in `internal/output/`
- Use `slog` for all logging (output to stderr)
- Output data to stdout only
- Environment variable prefix: `SMA_`
- New output formats should implement the `output.Formatter` interface
