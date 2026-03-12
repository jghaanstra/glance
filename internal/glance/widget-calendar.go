package glance

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"time"

	ics "github.com/arran4/golang-ical"
)

var calendarWidgetTemplate = mustParseTemplate("calendar.html", "widget-base.html")

var calendarWeekdaysToInt = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
}

type calendarWidget struct {
	widgetBase     `yaml:",inline"`
	FirstDayOfWeek string   `yaml:"first-day-of-week"`
	Ics            []string `yaml:"ics"`
	FirstDay       int      `yaml:"-"`
	cachedHTML     template.HTML `yaml:"-"`
	Events         string   `yaml:"-"`
}

type calendarEvent struct {
	Date time.Time
	Name string
}

func (widget *calendarWidget) initialize() error {
	widget.withTitle("Calendar").withError(nil)

	if widget.FirstDayOfWeek == "" {
		widget.FirstDayOfWeek = "monday"
	}

	if _, ok := calendarWeekdaysToInt[widget.FirstDayOfWeek]; !ok {
		return errors.New("invalid first day of week")
	}

	var icsEvents []*ics.VEvent
	for _, url := range widget.Ics {
		newEvents, err := ReadPublicIcs(url)
		if err != nil {
			log.Printf("calendar: failed to fetch ICS from %s: %v", url, err)
			continue
		}
		icsEvents = append(icsEvents, newEvents...)
	}

	var widgetEvents []calendarEvent
	for _, event := range icsEvents {
		startDate, err := event.GetStartAt()
		if err != nil {
			log.Printf("calendar: skipping event with invalid start date: %v", err)
			continue
		}
		summary := event.GetProperty("SUMMARY")
		if summary == nil {
			continue
		}
		widgetEvents = append(widgetEvents, calendarEvent{
			Date: startDate,
			Name: summary.Value,
		})
	}

	if widgetEvents != nil {
		jsonBytes, err := json.Marshal(widgetEvents)
		if err != nil {
			log.Printf("calendar: failed to marshal events: %v", err)
		} else {
			widget.Events = string(jsonBytes)
		}
	}

	widget.FirstDay = int(calendarWeekdaysToInt[widget.FirstDayOfWeek])
	widget.cachedHTML = widget.renderTemplate(widget, calendarWidgetTemplate)

	return nil
}

func (widget *calendarWidget) Render() template.HTML {
	return widget.cachedHTML
}

func ReadPublicIcs(url string) ([]*ics.VEvent, error) {
	response, err := http.Get(url) //nolint:gosec // URL comes from trusted config file
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	cal, err := ics.ParseCalendar(response.Body)
	if err != nil {
		return nil, err
	}
	return cal.Events(), nil
}
