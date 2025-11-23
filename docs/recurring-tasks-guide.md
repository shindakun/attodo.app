# Recurring Tasks

Never forget your regular tasks again! Recurring tasks in attodo automatically create a new task for you every time you complete one.

## What Are Recurring Tasks?

Recurring tasks are tasks that repeat on a regular schedule. When you complete a recurring task, attodo automatically creates a new one for you with the next due date. Perfect for:

- 📅 Regular meetings (weekly team standup)
- 💰 Monthly bills (rent, subscriptions)
- 🏃 Daily habits (exercise, meditation)
- 📧 Routine maintenance (change air filter every 3 months)
- 🎂 Annual events (birthdays, anniversaries)

## How It Works

1. **Create** a task and check the "Make this a recurring task 🔄" box
2. **Set the schedule** - choose how often it repeats (daily, weekly, monthly, yearly)
3. **Complete the task** - mark it done like any other task
4. **Next one appears** - a fresh task automatically appears with the next due date

That's it! The task keeps recreating itself every time you complete it.

## Creating a Recurring Task

### Step 1: Fill Out the Task Form

Start by creating a task normally:
- **Title**: What needs to be done
- **Description**: Optional details
- **Tags**: Optional tags for organization
- **Due Date**: When is the first occurrence due?

### Step 2: Enable Recurring

Check the box that says **"Make this a recurring task 🔄"**

This reveals the recurring options:

### Step 3: Choose Your Schedule

**Repeats:** Select how often the task recurs
- **Daily** - Every day or every N days
- **Weekly** - Specific days of the week
- **Monthly** - Same day each month
- **Yearly** - Once per year

**Every:** How many units between occurrences
- "1" = every single occurrence (default)
- "2" = every other occurrence
- "3" = every third occurrence, etc.

**On these days:** (Weekly tasks only)
- Check the days you want the task to appear
- Example: Mon, Wed, Fri for a 3-day-per-week workout

### Step 4: Submit

Click "Add Task" and you're done! Your recurring task is now active.

## Examples

### Daily Task
**"Review inbox"**
- Repeats: Daily
- Every: 1 day

Result: New task appears every single day after you complete it.

### Weekly Team Meetings
**"Team standup"**
- Repeats: Weekly
- Every: 1 week
- Days: Monday, Wednesday, Friday

Result: New task appears for the next Mon/Wed/Fri each time you complete it.

### Biweekly Tasks
**"Sprint planning"**
- Repeats: Weekly
- Every: 2 weeks
- Days: Monday

Result: New task appears every other Monday.

### Monthly Bills
**"Pay rent"**
- Repeats: Monthly
- Every: 1 month
- Due Date: Set to the 1st of the month

Result: New task appears on the 1st of the next month.

### Quarterly Reviews
**"Quarterly performance review"**
- Repeats: Monthly
- Every: 3 months

Result: New task appears 3 months after you complete each one.

### Annual Events
**"Mom's birthday"**
- Repeats: Yearly
- Every: 1 year
- Due Date: Set to her birthday

Result: New task appears on her birthday every year.

## Recognizing Recurring Tasks

Recurring tasks have a special **🔄 Recurring** badge next to their title. This helps you identify them in your task list.

## Completing Recurring Tasks

Just click "Mark Complete" like any other task!

When you complete a recurring task:
1. The current task is marked complete (and moves to your Completed tab)
2. A brand new task is instantly created
3. The new task has the next calculated due date
4. The new task appears in your Incomplete tasks

**Important:** The completed task doesn't disappear - it stays in your history so you can track what you've accomplished.

## How the Next Due Date is Calculated

### Daily Tasks
Simply adds N days to the current due date.

Example: Daily task due today (Nov 25)
- Complete it → Next one due Nov 26

### Weekly Tasks
Finds the next occurrence of the selected day(s).

Example: Weekly task on Mondays and Wednesdays, completed on Monday Nov 25
- Next one due Wednesday Nov 27

Example: Completed on Wednesday Nov 27
- Next one due Monday Dec 2

### Monthly Tasks
Same day next month (handles month-end correctly).

Example: Monthly task on the 15th, completed Nov 15
- Next one due Dec 15

Example: Monthly task on Jan 31, completed Jan 31
- Next one due Feb 28 (or 29 in leap years)

### Yearly Tasks
Same date next year.

Example: Annual task on Nov 25, completed Nov 25 2024
- Next one due Nov 25 2025

## Tips & Best Practices

### ✅ Do

- **Set a due date** - Recurring tasks need a starting due date
- **Be specific** - "Team standup" is better than "Meeting"
- **Use tags** - Tag recurring tasks for easy filtering
- **Start simple** - Begin with basic patterns, get fancy later

### ❌ Don't

- **Skip the due date** - Recurring tasks without due dates won't calculate correctly
- **Overthink it** - Start with the pattern that makes sense, adjust if needed
- **Delete by accident** - Deleting a recurring task removes the pattern

## Current Limitations

### Can't Edit the Pattern
Once you create a recurring task, you can't change its schedule through the UI yet.

**Workaround:** Delete the task and create a new one with the correct pattern.

### Only One at a Time
You only see the current occurrence, not future ones in advance.

This keeps your task list clean, but means you can't plan too far ahead.

### No Skip Button
Can't skip an occurrence without deleting it.

**Workaround:** If you need to skip one, just delete the current instance. The next one will still appear after you complete the following occurrence.

### No Exception Dates
Can't exclude specific dates like holidays.

**Workaround:** Delete the task on days you want to skip.

## Frequently Asked Questions

**Q: What happens if I delete a recurring task?**
A: The task and its recurring pattern are deleted. No future tasks will be created. Your completed instances remain in your history.

**Q: Can I have multiple recurring tasks with the same title?**
A: Yes! Each task is independent. You could have "Review email" as both a daily task and a weekly task.

**Q: Do recurring tasks work without a due date?**
A: They need an initial due date to calculate the next occurrence. The first one should have a due date set.

**Q: What if I don't complete a task before the next one is due?**
A: The system waits for you to complete the current one before creating the next. If you complete it late, the next occurrence calculates from the original due date (not when you completed it).

**Q: Can I pause a recurring task?**
A: Not directly. You can delete it to stop new occurrences, then recreate it later when you want to resume.

**Q: How many tasks can I make recurring?**
A: As many as you want! Each recurring task is independent.

**Q: Can I make an existing task recurring?**
A: Yes! When editing a task, check the "Make this a recurring task 🔄" option and set your schedule. Note: The pattern starts from when you save the edit, not retroactively.

## Coming Soon

We're working on improvements like:
- ✨ Edit recurring patterns without recreating
- 📅 See future occurrences before they're created
- ⏭️ Skip individual occurrences
- 🎯 Advanced patterns ("last Friday of the month")
- 🚫 Exception dates for holidays
- 📊 Completion streak tracking

## Need Help?

If something isn't working as expected:

1. **Check the badge** - Does the task show the 🔄 Recurring badge?
2. **Verify the due date** - Recurring tasks need a due date to work
3. **Contact support** - Reach out to [@attodo.app](https://bsky.app/profile/attodo.app) on Bluesky

---

**Happy recurring!** 🔄✨
