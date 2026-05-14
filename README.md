# croncheck

A lightweight daemon that monitors cron job execution and sends alerts when jobs fail or exceed expected durations.

## Installation

```bash
pip install croncheck
```

Or install from source:

```bash
git clone https://github.com/yourname/croncheck.git && cd croncheck && pip install .
```

## Usage

Wrap your cron command with `croncheck` to start monitoring it:

```bash
croncheck --job "backup" --timeout 300 -- /usr/local/bin/backup.sh
```

Define expected jobs in a simple config file:

```yaml
# croncheck.yml
jobs:
  backup:
    timeout: 300
    alert: slack
  db-cleanup:
    timeout: 60
    alert: email
```

Then run the daemon:

```bash
croncheck daemon --config croncheck.yml
```

When a job fails or runs longer than its defined timeout, croncheck fires an alert to your configured channel (Slack, email, webhook, etc.).

## Configuration

| Option | Description | Default |
|--------|-------------|---------|
| `--timeout` | Max allowed duration in seconds | `60` |
| `--job` | Unique job identifier | required |
| `--config` | Path to config file | `croncheck.yml` |

## License

MIT © 2024 yourname