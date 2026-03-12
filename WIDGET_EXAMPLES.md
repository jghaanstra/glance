# Widget Configuration Examples

## Analog Clock Widget

Displays one or more analog clock faces. The local time is always shown first; additional timezones can be added.

```yaml
- type: analog-clock
  title: World Clocks
  hide-am-pm-indicator: false
  hide-date: false
  dial-markers: NumericalFull   # NumericalFull | NumericalMinimal | None
  timezones:
    - timezone: Europe/Paris
      label: Paris
    - timezone: America/New_York
      label: New York
    - timezone: Asia/Tokyo
      label: Tokyo
```

**Features:**
- Smooth animated hands (hour, minute, second)
- Optional AM/PM indicator
- Optional date slot showing day and abbreviated month
- Three dial-marker styles: full numbers, minimal (3/6/9/12), or none
- Multiple timezone faces displayed side by side
- Client-side only — no server polling required

**Properties:**
- `hide-am-pm-indicator` (optional, default: false): Hide the AM/PM label inside the clock face
- `hide-date` (optional, default: false): Hide the day/month date slot
- `dial-markers` (optional, default: `NumericalFull`): Style of hour markers — `NumericalFull`, `NumericalMinimal`, or `None`
- `timezones` (optional): Array of additional timezone clocks
  - `timezone` (required): IANA timezone name (e.g. `America/New_York`)
  - `label` (optional): Display label; falls back to timezone name if omitted

---

## Calendar Widget (with ICS Events)

Displays a full-month navigation calendar with support for events from ICS (iCalendar) URLs. Event days are highlighted and show a tooltip on hover.

```yaml
- type: calendar
  first-day-of-week: monday
  ics:
    - https://example.com/calendar.ics
    - https://example.com/another.ics
```

**Features:**
- Full-month calendar view with previous/next month navigation
- Fetches events from one or more ICS/iCalendar URLs at startup
- Days with events are highlighted in the primary color
- Hover over an event day to see event names and times
- Configurable first day of the week

**Properties:**
- `first-day-of-week` (optional, default: `monday`): First day of the week. Options: `sunday`, `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`
- `ics` (optional): List of ICS calendar URLs. If omitted, no events are shown

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

---

## Background Refresh Configuration

The background refresh feature allows the server to proactively update widgets on a regular schedule,
which reduces perceived latency for end users. It is configured at the `server` level and applies to
all pages.

```yaml
server:
  host: 0.0.0.0
  port: 8080
  # how often the server wakes up to refresh widgets (every 5 minutes by default)
  background-refresh-interval: 5m
  # if true, refreshes happen only when there is at least one connected client
  background-refresh-only-when-clients: true
```

**How it works:**
1. The application starts a ticker that fires at the specified interval.
2. When the ticker fires, every widget on every active page is asked whether it
   will require an update within the next interval (`requiresUpdateWithin`).
3. Widgets that declare they need an update are refreshed in parallel in the
   background (their `update` method is called).
4. Client browsers still fetch widget data on page load, but most of the work is
   already done, resulting in faster page renders.

These options were introduced in the merge of PR #712 and are available in
**version 0.9 and later**.

---

## To-Do Widget (with Server-Side Storage)

Displays an interactive to-do list. Tasks can be stored either in the browser's local storage (default) or on the server, allowing tasks to sync across devices and browsers.

```yaml
- type: to-do
  id: my-tasks
  storage-type: server
```

**Browser storage (default):**

```yaml
- type: to-do
  id: work
```

**Server storage:**

```yaml
- type: to-do
  id: shopping
  storage-type: server
```

To use server storage, configure `data-path` in your server config:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  data-path: ./data  # directory where task files will be stored
```

Tasks are saved as JSON files at `<data-path>/to-do/<id>.json`.

> **Docker users:** Mount the data directory so data persists across container restarts:
> ```yaml
> volumes:
>   - /home/user/glance-data:/app/data
> ```
> And set `data-path: /app/data` in your config.

**Features:**
- Add, edit, check off, and delete tasks
- Drag-and-drop to reorder tasks
- Multiple independent lists using different `id` values
- Server storage syncs tasks across all devices/browsers
- Browser storage keeps tasks local to the device

**Properties:**
- `id` (optional): Unique identifier for the list. Used as the filename for server storage. Defaults to `"default"` when `storage-type` is `server`
- `storage-type` (optional, default: `browser`): Where tasks are stored
  - `browser` — browser localStorage, device/browser specific
  - `server` — JSON file on the server under `data-path/to-do/`


## Search Widget (with Suggestions & Shortcuts)

The search widget provides a customizable search bar. You can define
search `bangs` (shortcuts that redirect to a specific site using a
prefix), configure arbitrary `shortcuts` for quick navigation, and enable
`search suggestions` pulled from a suggestion engine.

```yaml
- type: search
  search-engine: duckduckgo          # base search URL, use {QUERY} placeholder if custom
  suggestions: true                  # enable suggestions dropdown
  suggestion-engine: google          # preset engine or custom URL
  shortcuts:
    - title: Gmail
      url: https://mail.google.com
      alias: gm
    - title: GitHub
      url: https://github.com
      alias: gh
  bangs:
    - title: YouTube
      shortcut: "!yt"
      url: https://www.youtube.com/results?search_query=!QUERY
  autofocus: true
  placeholder: "Type to search or use bangs…"
```

**Features:**
- Fuzzy-matched shortcuts with optional aliases
- Keyboard navigation of dropdown (Arrow keys, Tab/Shift+Tab, Enter)
- Search suggestions fetched from a configurable engine (Google, DuckDuckGo, Bing, Startpage, or custom)
- Bangs support (e.g. `!yt cats` to search YouTube)
- Detects URLs and navigates directly
- Visual error indicator when suggestion service fails

**Properties:**
- `search-engine` (required): Primary search URL, supports presets or custom URL containing `{QUERY}`
- `suggestions` (optional, default: false): Enable retrieval of real‑time suggestions
- `suggestion-engine` (optional): Preset name or custom URL for suggestions; defaults to the same as `search-engine` when suggestions enabled
- `shortcuts` (optional): Array of quick‑link shortcuts
  - `title` (required): Display text
  - `url` (required): Destination URL
  - `alias` (optional): Short alias for fuzzy matching
- `bangs` (optional): Array of search bangs
  - `title` (required): Display name
  - `shortcut` (required): Prefix string starting with `!`
  - `url` (required): Search URL containing `!QUERY` placeholder
- `autofocus`, `placeholder` behave as usual for input widgets


