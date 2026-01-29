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

## 📄&nbsp;Usage

```
taskseed sync --config config.yaml
```

Use `--dry-run` to see planned creations without writing. `taskseed doctor` validates credentials and calendar reachability.

## ⚙&nbsp;Configuration

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

[^reminders-rrule]: iOS Reminders supports recurring tasks created in its own UI but does not reliably consume externally provided `VTODO` items with `RRULE`, based on observed behavior and community reports.
