# AT Todo Features Guide

Comprehensive guide to all AT Todo features and how to use them effectively.

## Table of Contents

- [Quick Task Creation](#quick-task-creation)
- [Natural Language Date Parsing](#natural-language-date-parsing)
- [Hashtag Support](#hashtag-support)
- [Due Dates & Times](#due-dates--times)
- [Task Organization](#task-organization)
- [Notifications](#notifications)
- [Lists](#lists)
- [Progressive Web App](#progressive-web-app)

---

## Quick Task Creation

### Command Bar (Cmd+Shift+P)

The fastest way to create tasks is using the Command Bar.

**Opening the Command Bar:**
- **Mac**: Press `Cmd + Shift + P`
- **Windows/Linux**: Press `Ctrl + Shift + P`

![The command bar!](images/command-bar.png)

**Using the Command Bar:**

```
meeting tomorrow at 3pm, discuss project timeline #work
```

This single line creates a task with:
- **Title**: "meeting"
- **Due date**: Tomorrow at 3:00 PM
- **Description**: "discuss project timeline"
- **Tags**: ["work"]

**Syntax:**
```
[title] [date/time], [description] #tag1 #tag2
```

- Everything before the first comma becomes the title (with date extracted)
- Everything after the comma becomes the description
- Hashtags anywhere become tags
- Date/time expressions are automatically parsed and removed from title

**Examples:**
```
call client in 2 hours, discuss pricing #urgent #sales
review document next friday #work
buy groceries today #personal #shopping
```

---

## Natural Language Date Parsing

AT Todo understands natural language for dates and times. Just type naturally!

### Time-Based Relative Dates

Create tasks due at specific times from now:

| Input | Result |
|-------|--------|
| `in 2 hours` | 2 hours from current time |
| `in 30 minutes` | 30 minutes from now |
| `in one hour` | 1 hour from now |
| `in fifteen minutes` | 15 minutes from now |

**Examples:**
- `call back in 30 minutes`
- `meeting in two hours`
- `reminder 90 minutes from now`

### Day-Based Relative Dates

| Input | Result |
|-------|--------|
| `today` | Today at midnight |
| `tomorrow` | Tomorrow at midnight |
| `in 3 days` | 3 days from now |
| `next week` | 7 days from now |

### Specific Days

| Input | Result |
|-------|--------|
| `monday` | Next Monday |
| `next tuesday` | Tuesday of next week |
| `this friday` | This coming Friday |

### With Specific Times

Combine dates with times for precise scheduling:

| Input | Result |
|-------|--------|
| `tomorrow at 3pm` | Tomorrow at 3:00 PM |
| `monday at 9:30am` | Next Monday at 9:30 AM |
| `11/26 3:30pm` | November 26 at 3:30 PM |
| `next friday at 2pm` | Next Friday at 2:00 PM |

### Advanced Patterns

| Input | Result |
|-------|--------|
| `end of month` | Last day of current month |
| `end of week` | This coming Sunday |
| `next month` | Same date next month |
| `next quarter` | Start of next quarter |

### Timezone Support

All times are automatically converted to your local timezone:
- Displayed in your local time
- Stored in UTC in AT Protocol
- Notifications respect your timezone

---

## Hashtag Support

Add tags to tasks using hashtags - they're automatically extracted and removed from the text.

### Using Hashtags

**In the Command Bar:**
```
meeting tomorrow #work #urgent
```
Results in:
- Title: "meeting tomorrow"
- Tags: ["work", "urgent"]

**In Task Forms:**
Tags can also be added via the "Tags" field:
```
work, urgent, personal
```

### Tag Features

- **Max 10 tags per task**
- **Max 30 characters per tag**
- **Emoji supported**: `#🎯 #📧 #🚀`
- **Case insensitive**: `#Work` and `#work` are treated as the same
- **Auto-completion**: Suggests existing tags as you type
- **Clickable**: Click any tag to filter tasks by that tag

### Popular Tags

View your most-used tags in the dashboard sidebar to quickly filter tasks.

---

## Due Dates & Times

### Setting Due Dates

**Three ways to add due dates:**

1. **Natural language** (in title when creating):
   ```
   call client tomorrow at 3pm
   ```

2. **Command Bar**:
   ```
   review document next friday, needs approval #work
   ```

3. **Date picker** (when editing):
   - Use the date input for calendar selection
   - Use the time input for specific times
   - Leave blank for all-day tasks

### Editing Due Dates

When editing a task:
1. Click "Edit" on any task
2. Set date using the date picker
3. Optionally add a time
4. Or type in the title: `tomorrow at 3pm`

**The date/time inputs automatically show your local timezone values.**

### Visual Indicators

Tasks show visual indicators based on due dates:

- 🔴 **Overdue**: Tasks past their due date (red styling)
- 🟡 **Due Today**: Tasks due within 24 hours (yellow styling)
- 🔵 **Due Soon**: Tasks due within 3 days (blue styling)

### Due Date Display

Due dates are shown in friendly formats:
- "Today at 3:00pm"
- "Tomorrow at 9:30am"
- "Monday at 2:00pm"
- "Jan 15 at 4:00pm"
- "Yesterday" (for overdue)

---

## Task Organization

### Tags

Tags help categorize tasks:
- Add tags when creating: `#work #urgent`
- Add via tags field: `work, urgent, personal`
- Click tags to filter
- View popular tags in sidebar

### Lists

Organize related tasks into lists:

**Creating Lists:**
1. Navigate to Lists section
2. Click "Create New List"
3. Give it a name and optional description

**Adding Tasks to Lists:**
1. Click "Add to List" on any task
2. Select one or more lists
3. Tasks can belong to multiple lists

**List Features:**
- **Share lists**: Get a shareable link for any list
- **List views**: See all tasks in a list
- **Task counts**: See incomplete/complete counts per list

### Filtering

**Filter tasks by:**
- **Status**: Incomplete / Complete
- **Tags**: Click any tag to filter
- **Lists**: View tasks in specific lists
- **Due dates**: View overdue, today, or upcoming

---

## Notifications

AT Todo includes a comprehensive notification system to help you stay on top of tasks.

### In-App Notifications

**Notification Banner:**
- Appears at the top of the dashboard
- Shows counts of overdue, due today, and upcoming tasks
- Updates automatically
- Can be dismissed

**When you'll see it:**
- When you have overdue tasks (highest priority)
- When you have tasks due today
- When you have tasks due within 3 days

### Push Notifications

Get notified even when AT Todo isn't open.

**Enabling Push Notifications:**
1. Open Settings (gear icon)
2. Click "Enable Push Notifications"
3. Grant permission when prompted
4. Configure your preferences

**Notification Settings:**

**Timing Preferences:**
- ✅ Notify about overdue tasks
- ✅ Notify about tasks due today
- ⬜ Notify about tasks due within 3 days
- Set hours before due date (0-72 hours)

**Check Frequency:**
- Every 15 minutes
- Every 30 minutes (default)
- Every hour
- Every 2 hours

**Quiet Hours / Do Not Disturb:**
- Enable quiet hours mode
- Set start time (default: 10 PM)
- Set end time (default: 8 AM)
- No notifications during quiet hours

### Notification Grouping

AT Todo intelligently groups notifications to avoid spam:

**Single Task:**
```
Task Due Soon
"Call client" is due in 2 hours.
```

**Multiple Tasks:**
```
3 Tasks Due Today
• Call client (2h)
• Team meeting (4h)
• Review document (6h)
```

**Many Tasks:**
```
5 Overdue Tasks
• Submit report
• Follow up with vendor
• Update spreadsheet
...and 2 more
```

**Priority Order:**
1. Overdue tasks (shown first, require interaction)
2. Tasks due today (shown with times)
3. Tasks due soon (within 3 days)

**Only one notification shown at a time** to avoid overwhelming you.

### Smart Scheduling (Advanced)

AT Todo learns when you typically use the app and can optimize notification timing:

- **Usage tracking**: Records what hours you're most active
- **Optimal timing**: Sends notifications during your active hours
- **Privacy-first**: All tracking stored in your AT Protocol repository
- **Automatic**: No configuration needed

### Test Notifications

**Test your notification setup:**
1. Open Settings
2. Enable notifications if not already enabled
3. Click "Send Test Notification"
4. You should see a test notification appear

**If notifications aren't working:**
- Check browser notification permissions
- Ensure notifications aren't blocked in system settings
- Try a different browser (Chrome/Edge have best support)
- Check quiet hours settings

### Browser Compatibility

| Feature | Chrome | Firefox | Safari | Edge |
|---------|--------|---------|--------|------|
| Push Notifications | ✅ | ✅ | ✅ (iOS 16.4+) | ✅ |
| Background Sync | ✅ | ❌ | ❌ | ✅ |
| Periodic Sync | ✅ | ❌ | ❌ | ✅ |

**Note**: Safari and Firefox users will get notifications when the app is open or during periodic checks, but won't get background notifications.

---

## Lists

### Creating and Managing Lists

**Create a List:**
1. Navigate to the Lists tab
2. Click "Create New List"
3. Enter name and optional description
4. Click "Create List"

**Edit a List:**
1. Find the list in your Lists view
2. Click "Edit"
3. Update name or description
4. Click "Save"

**Delete a List:**
1. Click "Delete" on any list
2. Confirm deletion
3. Tasks remain - only list membership is removed

### Adding Tasks to Lists

**From a Task:**
1. Click "Add to List" on any task
2. Select one or more lists
3. Task is added to selected lists

**Tasks can belong to multiple lists** - useful for organization:
- Add "Submit report" to both "Work" and "High Priority" lists
- Add "Buy groceries" to "Personal" and "Weekly Routine" lists

### Viewing Lists

**List Overview:**
- See all your lists with task counts
- Shows incomplete and completed task counts
- Click list name to view all tasks in that list

**List Detail View:**
- See all tasks in the list
- Tasks show their full details
- Filter and sort tasks within the list
- Tasks indicate membership in other lists

### Sharing Lists

**Create a shareable link:**
1. Open any list
2. Click "Share List"
3. Copy the link
4. Share with anyone

**Shared lists are public** - anyone with the link can view tasks in that list (but not edit them).

---

## Progressive Web App

AT Todo is a Progressive Web App (PWA) that can be installed on your device.

### Installing AT Todo

**Desktop (Chrome/Edge):**
1. Visit AT Todo
2. Click the install icon in the address bar
3. Click "Install"

**iOS (Safari):**
1. Open AT Todo in Safari
2. Tap the Share button
3. Tap "Add to Home Screen"
4. Tap "Add"

**Android (Chrome):**
1. Open AT Todo in Chrome
2. Tap the menu (three dots)
3. Tap "Add to Home Screen"
4. Tap "Add"

### PWA Benefits

**Once installed:**
- 🚀 **Faster loading** - cached resources
- 📱 **Standalone app** - opens like a native app
- 🔔 **Better notifications** - system-level push notifications
- 📴 **Offline access** - view cached tasks without internet
- 💾 **Smaller footprint** - no app store download needed

### Offline Support

**What works offline:**
- View previously loaded tasks
- View cached lists
- See task details
- Navigate between views

**What requires internet:**
- Creating new tasks
- Editing tasks
- Deleting tasks
- Syncing with AT Protocol
- Push notifications

**Auto-sync when online** - changes sync automatically when connection is restored.

---

## Privacy & Data Ownership

### Your Data, Your Control

- **Decentralized storage**: Tasks stored in YOUR AT Protocol repository
- **No central database**: AT Todo servers don't store your tasks
- **Portable**: Your data works with any AT Protocol app
- **Private**: Only you can access your tasks (unless you share lists)

### Notification Privacy

- **Client-side logic**: All notification checking runs in your browser
- **No tracking**: No server-side tracking of notification views
- **Your settings**: Preferences stored in your AT Protocol repository
- **Local caching**: Settings cached locally for performance

### What's Stored Where

**In your AT Protocol repository:**
- Tasks (title, description, due dates, tags, status)
- Lists (name, description, task references)
- Notification settings (preferences, quiet hours)
- Usage patterns (for smart scheduling)

**On your device:**
- Session tokens (temporary)
- Cached tasks (for offline access)
- Service worker cache (for PWA functionality)

**Never stored:**
- Your password
- Browsing history
- Personal information beyond tasks

---

## Tips & Tricks

### Power User Tips

1. **Keyboard shortcuts**: `Cmd/Ctrl + Shift + P` for quick task creation
2. **Copy natural language**: Copy from calendar invites, paste into command bar
3. **Bulk tagging**: Add common tags to multiple related tasks
4. **Smart lists**: Create lists for contexts ("Work", "Home", "Errands")
5. **Time blocking**: Use due times to create a schedule
6. **Emoji tags**: Use `#🎯` for visual identification
7. **Morning review**: Check notification banner for today's priorities
8. **Weekly planning**: Use "due soon" filter to plan the week ahead

### Productivity Workflows

**Daily Review:**
1. Check notification banner for overdue/today tasks
2. Use command bar to quickly add new tasks
3. Mark completed tasks as done
4. Review "due soon" for tomorrow's priorities

**Weekly Planning:**
1. Review all incomplete tasks
2. Set due dates for the week
3. Use lists to organize by area (Work, Personal, etc.)
4. Enable "due soon" notifications for 3-day visibility

**Project Management:**
1. Create a list for each project
2. Tag tasks with project phases (#planning, #execution, #review)
3. Use due dates for milestones
4. Share project lists with team members

---

## Troubleshooting

### Common Issues

**Tasks not syncing:**
- Check internet connection
- Try refreshing the page
- Check AT Protocol status

**Notifications not appearing:**
- Check browser notification permissions
- Verify notifications enabled in Settings
- Test with "Send Test Notification"
- Check quiet hours settings
- Try a different browser

**Date parsing not working:**
- Check format: "tomorrow at 3pm" not "tomorrow 3pm"
- Use "at" before times
- Use AM/PM or 24-hour format
- Try date picker instead

**Timezone issues:**
- All times displayed in your local timezone
- Stored in UTC in AT Protocol
- Check system timezone settings
- Edit task to verify correct time shown

### Getting Help

- Check the [Getting Started guide](/docs/getting-started)
- Review this features guide
- Report issues on GitHub
- Contact via Bluesky

---

## What's Next?

AT Todo is continuously evolving. Planned features include:

- Recurring tasks
- Task templates
- Calendar integration
- Collaboration features
- Advanced filtering
- Custom views
- Mobile app

Stay tuned for updates!
