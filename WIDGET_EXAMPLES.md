# Widget Configuration Examples

## Calendar Widget (New - with Events)

The new calendar widget displays a 3-week calendar view with support for events from ICS (iCalendar) files.

```yaml
- type: calendar
  title: My Events Calendar
  icsurl: https://example.com/calendar.ics
  start-sunday: false
```

**Features:**
- Fetches events from ICS/iCalendar URLs
- Days with events are highlighted in the primary color
- Hover over event days to see event names
- Shows current month, week number, and year
- Updates on page load

**Properties:**
- `icsurl` (optional): URL to an ICS calendar file. If not provided, no events are displayed
- `start-sunday` (optional, default: false): If true, Sunday is the first day of the week

---

## Calendar (Legacy) Widget

The legacy calendar widget displays a simple 3-week calendar view without event support.

```yaml
- type: calendar-legacy
  start-sunday: false
```

**Properties:**
- `start-sunday` (optional, default: false): If true, Sunday is the first day of the week

---

## Count Timer Widget (Countdown/Reminder)

Displays a countdown timer to a specific date and time. Updates every second in real-time.

```yaml
- type: count-timer
  event-title: New Year 2026
  date: 2026-01-01T00:00:00Z

- type: count-timer
  title: Conference
  date: 2026-06-15T09:00:00-05:00
  href: https://conference.example.com
```

**Features:**
- Counts up or down to/from a target date
- Displays days, hours, minutes, and seconds
- Shows "FUTURE" status for future dates (primary color)
- Shows "PAST" status for past dates (highlight color)
- Optional link to an external URL
- Updates every second in the browser (no server load)

**Properties:**
- `date` (required): Target date in ISO 8601 format with timezone info
- `event-title` (optional): Event name. If set, title shows as "{event-title} ⋅ FUTURE/PAST"
- `title` (optional): Static title. Used when `event-title` is not set or empty
- `href` (optional): URL to link to when clicking the timer
