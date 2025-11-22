# AT Todo Lexicons

This directory contains the AT Protocol lexicon definitions for AT Todo.

## What are Lexicons?

Lexicons are schemas that define the structure of records stored in the AT Protocol. They specify what fields are required, their types, and validation rules.

## AT Todo Lexicons

AT Todo uses three lexicons to store data in users' personal data repositories:

### `app.attodo.task`

Individual todo tasks with optional due dates, tags, and completion status.

**Fields:**
- `title` (string, required, max 500 chars) - The task title
- `description` (string, optional, max 3000 chars) - Detailed description
- `completed` (boolean, required) - Whether the task is completed
- `createdAt` (datetime, required) - When the task was created
- `completedAt` (datetime, optional) - When the task was completed
- `dueDate` (datetime, optional) - When the task is due
- `tags` (array of strings, optional, max 10 tags, max 30 chars each) - User-defined tags

**Record Key:** `tid` (timestamp-based identifier)

### `app.attodo.list`

Collections of tasks organized into named lists.

**Fields:**
- `name` (string, required, max 100 chars) - The list name
- `description` (string, optional, max 500 chars) - List description
- `taskUris` (array of AT URIs, required) - References to tasks in this list
- `createdAt` (datetime, required) - When the list was created
- `updatedAt` (datetime, required) - When the list was last updated

**Record Key:** `tid` (timestamp-based identifier)

### `app.attodo.settings`

User preferences for notifications and UI settings. Single record per user.

**Fields:**
- `notifyOverdue` (boolean, default: true) - Notify for overdue tasks
- `notifyToday` (boolean, default: true) - Notify for tasks due today
- `notifySoon` (boolean, default: false) - Notify for tasks due within 3 days
- `hoursBefore` (integer, 0-72, default: 1) - Hours before due date to notify
- `checkFrequency` (integer, enum: [15, 30, 60, 120], default: 30) - Minutes between checks
- `quietHoursEnabled` (boolean, default: false) - Enable do-not-disturb mode
- `quietStart` (integer, 0-23, default: 22) - Quiet hours start hour
- `quietEnd` (integer, 0-23, default: 8) - Quiet hours end hour
- `pushEnabled` (boolean, default: false) - Browser push notifications enabled
- `taskInputCollapsed` (boolean, default: false) - Task input form collapsed by default
- `appUsageHours` (object, optional) - Usage pattern tracking for smart scheduling
- `updatedAt` (datetime, required) - Last update timestamp

**Record Key:** `literal:settings` (fixed key "settings")

## Publishing Lexicons

To publish these lexicons to the AT Protocol lexicon registry, you can use the `@atproto/lexicon` tools:

```bash
npm install -g @atproto/lexicon
lexicon publish ./lexicons/app/attodo/*.json
```

Alternatively, you can submit them to the [atproto.com lexicon catalog](https://atproto.com/lexicons).

## Data Ownership

All data stored using these lexicons lives in users' personal AT Protocol repositories. AT Todo servers never store user tasks, lists, or settings - they only facilitate reading and writing to users' own repositories.

This means:
- Users own their data
- Data is portable across AT Protocol apps
- No vendor lock-in
- Full transparency of what's stored

## Privacy

Tasks and lists are stored as public records in the AT Protocol repository, as specified in the protocol. Users should be aware that these records are publicly accessible via AT Protocol APIs.

Settings records contain user preferences and are also public by protocol design.
