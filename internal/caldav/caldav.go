// Package caldav wraps go-webdav helpers.
package caldav

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/justinrixx/retryhttp"
)

// Client handles CalDAV operations.
type Client struct {
	client       *caldav.Client
	calendarPath string
}

// Task represents an existing CalDAV VTODO.
type Task struct {
	UID        string
	Summary    string
	InstanceID string
	RuleID     string
	Occurrence string
	Completed  bool
}

// NewTask represents a VTODO to create.
type NewTask struct {
	UID        string
	Summary    string
	Notes      string
	Due        time.Time
	DateOnly   bool
	InstanceID string
	RuleID     string
	Occurrence string
	Timezone   string
}

const (
	taskseedIDProp   = "X-TASKSEED-ID"
	taskseedRuleProp = "X-TASKSEED-RULE"
	taskseedOccProp  = "X-TASKSEED-OCC"
)

// NewClient creates an authenticated CalDAV client.
func NewClient(endpoint, calendarURL, username, password string) (*Client, error) {
	baseClient := &http.Client{
		Transport: retryhttp.New(),
	}
	httpClient := webdav.HTTPClientWithBasicAuth(baseClient, username, password)
	calClient, err := caldav.NewClient(httpClient, endpoint)
	if err != nil {
		slog.Error("failed to create caldav client", "error", err)
		return nil, fmt.Errorf("create caldav client: %w", err)
	}

	path := calendarURL
	if u, err := url.Parse(calendarURL); err == nil && u.Path != "" {
		path = u.Path
	} else if err != nil {
		slog.Error("failed to parse calendar URL", "error", err)
		return nil, fmt.Errorf("parse calendar URL: %w", err)
	}

	return &Client{
		client:       calClient,
		calendarPath: path,
	}, nil
}

// QueryTasks fetches VTODO tasks in the provided time range.
func (c *Client) QueryTasks(ctx context.Context, start, end time.Time) ([]Task, error) {
	query := caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{
			Name: "VCALENDAR",
			Comps: []caldav.CalendarCompRequest{{
				Name:     "VTODO",
				AllProps: true,
			}},
		},
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{{
				Name:  "VTODO",
				Start: start,
				End:   end,
			}},
		},
	}

	objects, err := c.client.QueryCalendar(ctx, c.calendarPath, &query)
	if err != nil {
		slog.Error("failed to query caldav tasks", "calendar", c.calendarPath, "error", err)
		return nil, fmt.Errorf("query caldav tasks: %w", err)
	}

	var tasks []Task
	for _, obj := range objects {
		if obj.Data == nil || obj.Data.Component == nil {
			continue
		}

		for _, comp := range obj.Data.Component.Children {
			if comp.Name != ical.CompToDo {
				continue
			}
			tasks = append(tasks, calendarObjectToTask(comp))
		}
	}

	return tasks, nil
}

func calendarObjectToTask(comp *ical.Component) Task {
	uid, _ := comp.Props.Text(ical.PropUID)
	summary, _ := comp.Props.Text(ical.PropSummary)
	instanceID := textProp(comp, taskseedIDProp)
	ruleID := textProp(comp, taskseedRuleProp)
	occurrence := textProp(comp, taskseedOccProp)
	status, _ := comp.Props.Text(ical.PropStatus)
	completed := strings.EqualFold(status, "COMPLETED") || comp.Props.Get(ical.PropCompleted) != nil

	return Task{
		UID:        uid,
		Summary:    summary,
		InstanceID: instanceID,
		RuleID:     ruleID,
		Occurrence: occurrence,
		Completed:  completed,
	}
}

func textProp(comp *ical.Component, name string) string {
	prop := comp.Props.Get(name)
	if prop == nil {
		return ""
	}
	return prop.Value
}

// CreateTask writes a new task resource to the calendar.
func (c *Client) CreateTask(ctx context.Context, task NewTask) error {
	todo := ical.NewComponent(ical.CompToDo)

	todo.Props.SetText(ical.PropUID, task.UID)
	todo.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	todo.Props.SetText(ical.PropSummary, task.Summary)
	if task.Notes != "" {
		todo.Props.SetText(ical.PropDescription, task.Notes)
	}
	if task.DateOnly {
		prop := ical.NewProp(ical.PropDue)
		prop.SetDate(task.Due)
		todo.Props.Set(prop)
	} else {
		prop := ical.NewProp(ical.PropDue)
		prop.SetDateTime(task.Due)
		if task.Timezone != "" {
			prop.Params.Set(ical.ParamTimezoneID, task.Timezone)
		}
		todo.Props.Set(prop)
	}

	todo.Props.SetText(taskseedIDProp, task.InstanceID)
	todo.Props.SetText(taskseedRuleProp, task.RuleID)
	todo.Props.SetText(taskseedOccProp, task.Occurrence)

	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//taskseed//EN")
	cal.Children = append(cal.Children, todo)

	resource := strings.TrimSuffix(task.InstanceID, ".ics")
	resource += ".ics"

	_, err := c.client.PutCalendarObject(ctx, joinPath(c.calendarPath, resource), cal)
	if err != nil {
		slog.Error("failed to create caldav task", "calendar", c.calendarPath, "id", task.InstanceID, "error", err)
		return fmt.Errorf("create caldav task: %w", err)
	}

	return nil
}

func joinPath(base, name string) string {
	if strings.HasSuffix(base, "/") {
		return base + name
	}
	return base + "/" + name
}
