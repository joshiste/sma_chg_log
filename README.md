# SMA EV Charging Log

A CLI tool to fetch and export EV charging events/sessions from your SMA EV Charger with ennexOS to JSON, CSV or PDF.

## Notice

This project is not affiliated with SMA Solar Technology AG in any way. Use at your own risk.

The SMA, SMA ennexOS names and logos are trademarks of SMA Solar Technology AG.

## Features

- Fetches charging events from SMA EV Charger API and pairs charging start/stop events into sessions
- Exports all charging sessions to JSON, CSV, or PDF
- Filter by month
- Map authentication IDs to user-friendly names

## Installation

### Binary releases

Download the latest release from the [releases page](https://github.com/joshiste/sma_chg_log/releases).

### Docker

```bash
docker pull ghcr.io/joshiste/sma_chg_log:latest
```

### From source

```bash
go install github.com/joshiste/sma_chg_log@latest
```

## Usage

```bash
# PDF report with all charging sessions for December 2025
sma_chg_log --host device.local --username admin --password yourpassword --format pdf --month 2026-01 --output report-2026-01.pdf

# PDF report with mapping the authentication 
sma_chg_log --host device.local --username admin --password yourpassword --format pdf --month 2026-01 --output report-2026-01.pdf --map-authentication "old-auth-value:new-auth-value" --map-authentication ":value-for-missing-auth"

# CSV export with all charging sessions
sma_chg_log --host device.local --username admin --password yourpassword --format csv --output report-2026-01.csv
```

### Docker

```bash
docker run --rm \
  ghcr.io/joshiste/sma_chg_log:latest \
  --host device.local --username admin --password yourpassword --format csv
```

## Commands

### sessions (default)

Fetches charging events and outputs paired charging sessions. This is the default command when none is specified.

```bash
sma_chg_log sessions --host device.local --username admin --password secret
```

**Options:**
- `--map-authentication` - Map authentication values (format: `old:new`, can be specified multiple times)
  - Use empty old value to set default: `--map-authentication ":Unknown User"`

**Supported formats:** json, csv, pdf

### events

Fetches and outputs raw charging event messages (start/stop) without pairing.

```bash
sma_chg_log events --host device.local --username admin --password secret
```

**Supported formats:** json only

## Configuration

All parameters can be set via command line flags or environment variables. Flags take precedence.

### Global Flags

| Parameter | Flag              | Environment Variable | Required | Description                             |
|-----------|-------------------|----------------------|----------|-----------------------------------------|
| Host      | `-h, --host`      | `SMA_HOST`           | Yes      | SMA device hostname (defaults to https) |
| Username  | `-u, --username`  | `SMA_USERNAME`       | Yes      | Authentication username                 |
| Password  | `-p, --password`  | `SMA_PASSWORD`       | Yes      | Authentication password                 |
| Format    | `-f, --format`    | `SMA_FORMAT`         | No       | Output: json, csv, pdf (default: json)  |
| Output    | `-o, --output`    | `SMA_OUTPUT`         | No       | Output file (default: `-` for stdout)   |
| Month     | `-m, --month`     | `SMA_MONTH`          | No       | Filter by month (YYYY-MM)               |
| Log Level | `-l, --log-level` | `SMA_LOG_LEVEL`      | No       | trace, debug, info, warn, error         |

### Sessions Command Flags

| Parameter            | Flag                        | Description                                      |
|----------------------|-----------------------------|--------------------------------------------------|
| Map Authentication   | `-a, --map-authentication`  | Map auth values (format: `old:new`, repeatable)  |

## Output Formats

### JSON Lines
One JSON object per charging/session event per line.

### CSV
Paired charging sessions with columns: record date, charger name, authentication, start time, end time, consumption (kWh).

### PDF
Same as CSV with a summary showing total records and consumption.

## License

MIT License - see [LICENSE](LICENSE) file.
