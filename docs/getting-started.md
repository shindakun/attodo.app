# Getting Started

This guide will help you get started with AT Todo and create your first task.

## Step 1: Login with Bluesky

AT Todo uses your existing Bluesky account for authentication.

1. Visit the AT Todo homepage
2. Enter your Bluesky handle (e.g., `alice.bsky.social`)
3. Click "Login with Bluesky"
4. You'll be redirected to authorize the application
5. Grant AT Todo permission to manage your tasks

**Note:** AT Todo only requests permission to store task data in your repository. It does not access your social media posts or followers. The auth screen is needlessly scary currently. Thanks Bluesky!

## Step 2: Create Your First Task

Once logged in, you'll see the dashboard with the task creation form.

1. **Enter a title** - A brief description of what needs to be done
   - Example: "Finish project proposal"

2. **Add a description (optional)** - Additional details about the task
   - Example: "Include budget breakdown and timeline"

3. **Click "Add Task"** - Your task will be created and appear in the list

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
2. Update the title or description
3. Click "Save" to update the task
4. Click "Cancel" to discard changes

### Delete Tasks

To permanently remove a task:

1. Click the "Delete" button
2. Confirm the deletion
3. The task will be removed from your repository

**Warning:** Deleted tasks cannot be recovered!

## Step 4: Using Tabs

AT Todo organizes your tasks into two tabs:

- **Incomplete** - Tasks that are still pending
- **Completed** - Tasks you've finished

Click the tab buttons to switch between views. This helps you focus on what needs to be done while keeping a record of completed work.

## Step 5: Install as PWA (Optional)

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

## Step 6: Working Offline

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

- Start organizing your tasks effectively
- Explore the [Technical Details](/docs/technical) to learn how AT Todo works
- Check the [main documentation](/docs) for more information

Happy task managing! 🎯
