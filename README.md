# taskseed

taskseed materializes recurring tasks into a CalDAV task list. It expands declarative rules into concrete VTODO entries so clients that lack reliable recurrence support still receive upcoming tasks.

## Usage

```
taskseed sync --config config.yaml
```

Use `--dry-run` to see planned creations without writing. `taskseed doctor` validates credentials and calendar reachability.

## Configuration

Find below an example YAML file. By default, we read `config.yaml` from the current working directory; override with `--config` or `TASKSEED_CONFIG`.

For **credentials**, set the environment variables `TASKSEED_CALDAV_USERNAME` and `TASKSEED_CALDAV_PASSWORD`.

```yaml
server:
  url: https://cal.example.com
target:
  url: https://cal.example.com/dav/calendars/user/tasks/
sync:
  horizonDays: 365
  lookbackDays: 7
defaults:
  timezone: UTC
  due:
    time: "09:00"
rules:
  - id: water_plants
    title: Water the plants
    schedule:
      kind: weekly
      weekdays: [monday, thursday]
```
