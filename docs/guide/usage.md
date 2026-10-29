# Monitoring and themes

## Navigation and status filters

Click the UptimeLit logo to return to your monitors. The **+** icon in the navbar opens the new-monitor form; the bell opens your notification channels. Monitor cards open their actions through the three-dot menu.

The status badges above the cards are filters:

| Badge | Shows |
| --- | --- |
| **All** | Every monitor, including pending checks; clears the status filter and search |
| **Up** | Unpaused monitors whose last check succeeded |
| **Down** | Unpaused monitors whose last check failed |
| **Paused** | Monitors that are paused, regardless of their last result |

The active filter is highlighted. Badge counts describe all your monitors, independently of the search query. Search matches monitor names and URLs/addresses and combines with the selected status filter.

**All** shows the overall monitor count. If no matches remain, use **All** or **Clear filters** to reset the view.

## HTTP and TCP checks

HTTP checks try `HEAD` first. A `405 Method Not Allowed` response causes a retry with `GET`, so API health endpoints that only support GET can still be monitored. Responses with status codes `200`–`399` count as UP. Redirects are reported without following them. Network errors, TLS validation failures and other status codes count as DOWN. The HTTP client timeout is 10 seconds.

TCP monitors attempt to open a connection to `host:port`; a successful connection counts as UP. They do not send application-level health queries, and their cards have no HTTP status code.

Choose a check interval of at least 10 seconds. The form defaults to 60 seconds and a failure threshold of three. The threshold controls outage alerts; it does not delay the displayed status change. Workers stagger startup briefly. An outage alert is sent once when consecutive failures reach the threshold, including failures from startup. A recovery alert is sent only after an outage alert was triggered; brief failures below the threshold produce neither notification.

## Cards and history

A card shows its name/address, current state, check type/interval, recent check bars, 24-hour uptime, last response time, HTTP status when applicable and the number of linked notifiers. The monitor list refreshes approximately every five seconds while the tab is visible. The 24-hour aggregate is refreshed separately in the background.

Green bars indicate successful checks; red bars indicate failures. Empty bars mean no recorded result for those positions. Paused monitors use muted bars. Hover a recorded bar to enlarge it with a short animation and show a tooltip with its timestamp, status, protocol, HTTP code where applicable and response time. The timeline uses one shared tooltip that follows the active check. Hover areas meet without gaps, and only the check shown in the tooltip is enlarged. You can move the pointer into the tooltip to read it. Keyboard focus also opens the tooltip; arrow keys and Home/End move between checks, and Escape dismisses it. On touch screens, use **View history** to see check details. Reduced-motion preferences disable the animation.

Click **View history** below the bars or choose **History** from the actions menu to inspect the most recent 100 checks. The dashboard shows up to 50 recent check bars. Check history is retained for 90 days with periodic cleanup.

Choose **Edit** to change the monitor's name, type, target, interval, threshold or attached notifiers. **Pause** stops its worker; **Resume** starts it again. **Delete** removes the monitor and its history after confirmation.

## Themes

Open the theme control in the navbar, the mobile menu or the floating button on the login page. Built-in choices are System, Crema, Latte, Frappé, Macchiato, Mocha and Black. System resolves to Crema or Frappé according to the browser's light/dark preference.

**Edit colors…** opens a custom theme editor for background, tiles, text, primary, accent, danger and success colors. You can choose a light/dark scheme and optionally provide a background image URL. Save it to make **Custom** available in the theme menu; **Reset** removes the custom theme and returns to System.

Theme preferences are stored in the current browser, not synchronized with the user account. Fonts and CSS are bundled with the application. See the [screenshots](screenshots.md) for light and dark examples.
