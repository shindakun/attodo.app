# Getting Started

This guide will help you get started with AT Todo and create your first task.

**Good news!** AT Todo is completely free to use with all features included. We also offer an optional Supporter tier ($24/year) for those who want to help support development and server costs. See our [pricing details](/docs#pricing) for more information.

## Step 1: Login with Bluesky

AT Todo uses your existing Bluesky account for authentication.

1. Visit the AT Todo homepage
2. Enter your Bluesky handle (e.g., `alice.bsky.social`)
3. Click "Login with Bluesky"
4. You'll be redirected to authorize the application
5. Grant AT Todo permission to manage your tasks

**Note:** AT Todo only requests permission to store task data in your repository. It does not access your social media posts or followers. The auth screen is needlessly scary currently. Thanks Bluesky!

## Step 2: Create Your First Task

Once logged in, you have multiple ways to create tasks:

### Quick Method: Command Bar

The fastest way to create tasks is using the command bar:

1. Press **Cmd+Shift+P** (Mac) or **Ctrl+Shift+P** (Windows/Linux)
2. Type your task naturally: `call client tomorrow at 3pm, discuss pricing #work`
3. Press Enter
4. Task is created instantly!

The command bar understands:
- **Natural dates**: "tomorrow", "next friday", "in 2 hours"
- **Times**: "at 3pm", "9:30am"
- **Hashtags**: #work, #urgent (automatically become tags)
- **Descriptions**: Everything after a comma

### Traditional Method: Task Form

1. **Enter a title** - A brief description of what needs to be done
   - Example: "Finish project proposal"

2. **Add a description (optional)** - Additional details about the task
   - Example: "Include budget breakdown and timeline"

3. **Set a due date (optional)** - When the task should be completed
   - Use the date picker or type naturally in the title
   - Add a time if needed

4. **Add tags (optional)** - Organize your task with comma-separated tags
   - Example: "work, urgent, client-meeting"

5. **Click "Add Task"** - Your task will be created and appear in the list

### Natural Language Examples

You can type dates and times naturally in the title:

- `submit report tomorrow`
- `meeting next friday at 2pm`
- `call back in 30 minutes`
- `review document 11/26 at 3:30pm`

AT Todo will automatically:
- Extract the date and time
- Set the due date on the task
- Clean up the title

## Step 3: Managing Tasks

### Mark Tasks as Complete

Click the "Mark Complete" button on any task to mark it as done. The task will:
- Move to the "Completed" tab
- Be timestamped with the completion time
- Remain in your task list for reference

### Mark Tasks as Incomplete

In the "Completed" tab, click "Mark Incomplete" to move a task back to the incomplete list. This is useful if you need to revisit a task.

### Edit Tasks

Need to fix a typo or add more details?

1. Click the "Edit" button on any task
2. Update the title, description, or tags
3. Click "Save" to update the task
4. Click "Cancel" to discard changes

### Delete Tasks

To permanently remove a task:

1. Click the "Delete" button
2. Confirm the deletion
3. The task will be removed from your repository

**Warning:** Deleted tasks cannot be recovered!

## Step 4: Using Tags to Organize

Tags are a powerful way to organize and categorize your tasks.

### Adding Tags

When creating or editing a task:
1. Enter tags separated by commas in the tags field
2. Example: `work, urgent, meeting`
3. Tags will appear as clickable badges on the task

### Filtering by Tags

Click any tag badge to filter your tasks:
- Only tasks with that tag will be shown
- Both incomplete and completed tasks are filtered
- Click "Clear Filter" to see all tasks again

### Popular Tags Widget

The dashboard shows your most-used tags:
- Tags are sorted by frequency
- Click any tag to filter by it
- Tag counts update automatically

### Tag Autocomplete

When typing tags, you'll see suggestions from your existing tags:
- Start typing to see matching tags
- Click a suggestion to use it
- Helps maintain consistent tag naming

## Step 5: Using Tabs

AT Todo organizes your tasks into tabs:

- **Incomplete** - Tasks that are still pending
- **Completed** - Tasks you've finished
- **Lists** - Organized collections of tasks

Click the tab buttons to switch between views. This helps you focus on what needs to be done while keeping a record of completed work.

## Step 6: Enable Notifications (Optional)

Stay on top of your tasks with smart notifications.

### Setting Up Notifications

1. Click the **Settings** icon (gear) in the dashboard
2. Click **"Enable Push Notifications"**
3. Grant permission when your browser prompts you
4. Configure your notification preferences

### Notification Options

- **Overdue tasks**: Get notified about tasks past their due date
- **Due today**: Alerts for tasks due within 24 hours
- **Due soon**: Reminders for tasks due within 3 days
- **Check frequency**: How often to check (15 min to 2 hours)
- **Quiet hours**: Set do-not-disturb times (e.g., 10 PM - 8 AM)

### Types of Notifications

**In-App Banner:**
- Appears at top of dashboard
- Shows overdue and upcoming task counts
- Updates automatically

**Push Notifications:**
- Work even when AT Todo is closed
- Smart grouping to avoid spam
- Click to open AT Todo

**Example notifications:**
- "Task Due Soon: 'Call client' is due in 2 hours"
- "3 Tasks Due Today" (with list of tasks)
- "5 Overdue Tasks" (grouped to avoid spam)

## Step 7: Install as PWA (Optional)

AT Todo works as a Progressive Web App, which means you can install it on your device:

### On Desktop (Chrome/Edge)
1. Look for the install icon in the address bar
2. Click "Install AT Todo"
3. The app will open in its own window

### On Mobile (iOS/Android)
1. Open the browser menu
2. Select "Add to Home Screen"
3. The app will appear as an icon on your home screen

### Benefits of Installing
- Quick access from your device
- Native app-like experience
- Better notification support
- Faster loading with offline caching

## Step 8: Working Offline

AT Todo includes offline support:

- Recently viewed tasks are cached
- You can browse tasks without internet
- Changes made offline will sync when reconnected (I think hasn't been tested very well)

**Note:** Creating new tasks requires an internet connection to store them in your AT Protocol repository.

## Tips for Effective Task Management

### Keep Titles Concise
Use short, action-oriented titles that clearly describe the task.

✅ Good: "Schedule dentist appointment"
❌ Too long: "I need to remember to call the dentist office and schedule an appointment for next month"

### Use Descriptions for Details
Save additional information in the description field:
- Deadlines
- Sub-steps
- Related links or references
- Notes and context

### Organize with Tags
Use tags to categorize and group related tasks:
- **By project**: `client-a`, `website-redesign`, `quarterly-report`
- **By priority**: `urgent`, `high-priority`, `low-priority`
- **By context**: `home`, `work`, `errands`, `phone-calls`
- **By time**: `today`, `this-week`, `someday`

**Tip**: Keep tag names consistent and avoid creating too many similar tags (e.g., use `work` instead of both `work` and `office`).

### Review Regularly
- Check your incomplete tasks daily
- Review completed tasks weekly
- Delete old completed tasks to keep your list manageable

### Break Down Large Tasks
Instead of "Plan vacation", create separate tasks:
- Research destinations
- Book flights
- Reserve hotel
- Plan activities

## Troubleshooting

### Can't Login?

- Verify your Bluesky handle is correct
- Make sure you're using the full handle (e.g., `alice.bsky.social`)
- Check your internet connection
- Try clearing browser cookies and logging in again

### Tasks Not Loading?

- Refresh the page
- Check your internet connection
- Log out and log back in
- Verify you're using the same account

### Session Expired?

If you see a "session expired" message:
1. You'll be redirected to the login page
2. Enter your handle again
3. Authorize the app
4. Your tasks will reload

Sessions expire for security after a period of inactivity.

## Next Steps

Now that you're familiar with the basics, you can:

- Master the **command bar** (Cmd+Shift+P) for ultra-fast task creation
- Learn about **natural language date parsing** in the [Features Guide](/docs/features)
- Set up **smart notifications** to never miss a deadline
- Create **lists** to organize related tasks
- Read the [Features Guide](/docs/features) for advanced tips and tricks
- Check the [main documentation](/docs) for more information

Happy task managing! 🎯
