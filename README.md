<div align="center">
	<h1>taskseed</h1>
	<h4 align="center">
		Recurring CalDAV tasks that actually work.
	</h4>
	<p>
		<strong>taskseed</strong> injects recurring tasks into your CalDAV task list, so every client sees upcoming tasks (VTODOs).
	</p>
</div>

<p align="center">
	<a href="https://github.com/eikendev/taskseed/actions"><img alt="Build status" src="https://img.shields.io/github/actions/workflow/status/eikendev/taskseed/main.yml?branch=main"/></a>&nbsp;
	<a href="https://github.com/eikendev/taskseed/blob/main/LICENSE"><img alt="License" src="https://img.shields.io/github/license/eikendev/taskseed"/></a>&nbsp;
</p>

## ✨&nbsp;Why taskseed?

Recurring `VTODO` tasks are handled inconsistently across CalDAV clients. Some ignore recurrence rules entirely.

**taskseed** avoids this by expanding recurring tasks on the server side. Instead of relying on clients to interpret recurrence metadata, it materializes upcoming tasks as concrete `VTODO` entries.

## 🧠&nbsp;How it works

- You declare your recurring tasks once in a simple configuration file
- taskseed runs on the server and regularly generates upcoming task instances (`VTODO` items)
- Each occurrence is written as a normal, single task into your CalDAV task list
- Your apps simply show upcoming tasks; no special support or setup for recurrence required

## 🚀&nbsp;Installation

### Install directly via Go

```bash
go install github.com/eikendev/taskseed/cmd/taskseed@latest
```

### Download pre-built release

Grab the latest release binary for your platform from the [GitHub Releases page](https://github.com/eikendev/taskseed/releases/latest).


## 📄&nbsp;Usage

### ⚙&nbsp;Configuration

Find below an example YAML file. By default, we read `config.yaml` from the current working directory; override with `--config` or `TASKSEED_CONFIG`.

For **credentials**, set the environment variables `TASKSEED_CALDAV_USERNAME` and `TASKSEED_CALDAV_PASSWORD`.

> [!IMPORTANT]
> Rule IDs must be stable over time. taskseed uses the rule ID to derive a deterministic instance ID (`X-TASKSEED-ID`) for each occurrence. If you change a rule ID, taskseed treats it as a brand-new rule and will create a new series of tasks, leaving the old series behind.

```yaml
server:
  # Base CalDAV endpoint (required)
  url: https://cal.example.com/remote.php/dav

target:
  # Full URL to the target task list (required)
  url: https://cal.example.com/dav/calendars/user/tasks/

sync:
  # How far into the future to materialize tasks (required; > 0)
  horizonDays: 365
  # How far into the past to scan for existing tasks (required; > 0)
  lookbackDays: 7

defaults:
  # IANA timezone for task generation (optional; default: UTC)
  timezone: UTC
  due:
    # Due time for created tasks (optional; 24h HH:MM)
    time: "09:00"
    # If true, write due dates as date-only values (optional; default: false)
    dateOnly: false

rules:
  # Each rule defines one recurring task (required)
  - id: water_plants
    # Stable identifier used for deduplication (required; unique)
    title: Water the houseplants
    # Task description (optional)
    notes: "Check top inch of soil; skip if still moist."
    schedule:
      kind: weekly
      # Runs on specific weekdays each week.
      # List of weekdays to include (required for weekly)
      weekdays: [monday, thursday]

  - id: take_vitamins
    title: Take vitamins
    schedule:
      # Runs every N days from the first occurrence.
      # Integer interval in days (required)
      kind: every_n_days
      everyNDays: 2

  - id: change_sheets
    title: Change bedsheets
    schedule:
      # Runs on specific days of the month.
      # List of month days (required; 1–31)
      kind: monthly_day
      monthDays: [1]

  - id: call_parents
    title: Call parents
    schedule:
      # Runs on the Nth weekday of each month.
      # Occurrence number in the month (required)
      # Weekday name (required)
      kind: monthly_nth_weekday
      nth: 1
      nthWeekday: sunday

  - id: renew_subscription
    title: Renew subscriptions
    schedule:
      # Runs on a specific month/day each year.
      # Month number (required; 1–12)
      # Day of month (required; 1–31)
      kind: yearly_date
      month: 12
      day: 1

  - id: dental_checkup
    title: Schedule dental checkup
    schedule:
      # Runs on the Nth weekday of a specific month.
      # Month number (required; 1–12)
      # Occurrence number in the month (required)
      # Weekday name (required)
      kind: yearly_nth_weekday
      month: 6
      nth: 2
      yearlyNthWeekday: monday
```

### Run

```bash
taskseed sync
```

Run in dry-run mode to see what would happen without making changes:

```bash
taskseed sync --dry-run
```

Increase logging verbosity:

```bash
taskseed sync --verbose
```

Use a different configuration file:

```bash
taskseed sync --config /path/to/config.yaml
```

Validate configuration and connectivity:

```bash
taskseed doctor
```
