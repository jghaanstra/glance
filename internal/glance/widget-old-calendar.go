package glance

import (
	"context"
	"html/template"
	"time"
)

var oldCalendarWidgetTemplate = mustParseTemplate("old-calendar.html", "widget-base.html")

type oldCalendarWidget struct {
	widgetBase  `yaml:",inline"`
	Calendar    *oldCalendarData
	StartSunday bool `yaml:"start-sunday"`
}

func (widget *oldCalendarWidget) initialize() error {
	widget.withTitle("Calendar").withCacheOnTheHour()

	return nil
}

func (widget *oldCalendarWidget) update(ctx context.Context) {
	widget.Calendar = newOldCalendar(time.Now(), widget.StartSunday)
	widget.withError(nil).scheduleNextUpdate()
}

func (widget *oldCalendarWidget) Render() template.HTML {
	return widget.renderTemplate(widget, oldCalendarWidgetTemplate)
}

type oldCalendarData struct {
	CurrentDay        int
	CurrentWeekNumber int
	CurrentMonthName  string
	CurrentYear       int
	Days              []int
}

// TODO: very inflexible, refactor to allow more customizability
// TODO: allow changing between showing the previous and next week and the entire month
func newOldCalendar(now time.Time, startSunday bool) *oldCalendarData {
	year, week := now.ISOWeek()
	weekday := now.Weekday()
	if !startSunday {
		weekday = (weekday + 6) % 7 // Shift Monday to 0
	}

	currentMonthDays := oldDaysInMonth(now.Month(), year)

	var previousMonthDays int

	if previousMonthNumber := now.Month() - 1; previousMonthNumber < 1 {
		previousMonthDays = oldDaysInMonth(12, year-1)
	} else {
		previousMonthDays = oldDaysInMonth(previousMonthNumber, year)
	}

	startDaysFrom := now.Day() - int(weekday) - 7

	days := make([]int, 21)

	for i := 0; i < 21; i++ {
		day := startDaysFrom + i

		if day < 1 {
			day = previousMonthDays + day
		} else if day > currentMonthDays {
			day = day - currentMonthDays
		}

		days[i] = day
	}

	return &oldCalendarData{
		CurrentDay:        now.Day(),
		CurrentWeekNumber: week,
		CurrentMonthName:  now.Month().String(),
		CurrentYear:       year,
		Days:              days,
	}
}

func oldDaysInMonth(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
