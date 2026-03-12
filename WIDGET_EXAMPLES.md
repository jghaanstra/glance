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

---

## Bookmarks Widget (with Automatic Favicons)

Displays bookmarks organized in groups with automatic favicon support.

```yaml
- type: bookmarks
  groups:
    - title: Development
      links:
        - title: GitHub
          url: https://github.com
          icon: favicon
        - title: Stack Overflow
          url: https://stackoverflow.com
          icon: favicon
    - title: Social
      links:
        - title: Twitter
          url: https://twitter.com
          icon: https://twitter.com/favicon.ico
        - title: Reddit
          url: https://reddit.com
          icon:
            type: emoji
            value: "🔗"
```

**Features:**
- Organize bookmarks into groups
- Support for automatic favicons using `icon: favicon`
- Manual favicon URLs supported
- Emoji icons and other icon types
- Customize target behavior (same tab / new tab)
- Hide arrow icons optionally

**Automatic Favicon Setup:**

To use automatic favicons, you must configure a `cache-path` in your server config:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  cache-path: /var/cache/glance  # Path where favicons will be cached
```

Then in bookmarks, simply use `icon: favicon`:

```yaml
- type: bookmarks
  groups:
    - title: My Links
      links:
        - title: GitHub
          url: https://github.com
          icon: favicon  # Automatically fetches and caches favicon
```

**Features:**
- Automatically fetches favicons from domains via Google's favicon service
- Caches favicons locally for 1 week
- Falls back to stale cache if fetch fails
- No external HTTP requests needed after caching
- Works with any URL

**Properties:**
- `groups` (required): Array of bookmark groups
- `title` (optional): Group title
- `color` (optional): Highlight color for group
- `same-tab` (optional, default: false): Open links in same tab
- `hide-arrow` (optional, default: false): Hide external link arrow
- `target` (optional): Override HTML target attribute
- `links` (required): Array of bookmark links
  - `title` (required): Link title
  - `url` (required): Link URL
  - `icon` (optional): Icon (supports `favicon`, emoji, URLs, or icon libraries)
  - `description` (optional): Link description

