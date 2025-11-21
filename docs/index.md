# Documentation

Welcome to AT Todo documentation! AT Todo is a decentralized todo application built on the AT Protocol, allowing you to own and control your task data.

## What is AT Todo?

AT Todo leverages the AT Protocol (the protocol behind Bluesky) to store your tasks in your own personal data repository. This means:

- **You own your data** - Tasks are stored in your AT Protocol repository
- **Decentralized** - No central server owns your information
- **Portable** - Your data can be accessed by any AT Protocol-compatible application

## Features

### Core Task Management
- ✅ Create and manage tasks with titles and descriptions
- ✅ Mark tasks as complete/incomplete
- ✅ Edit and delete tasks
- ✅ Set due dates with specific times
- ✅ Organize tasks with tags
- ✅ Filter tasks by tags and status

### Quick Task Creation
- ✅ **Command Bar** (Cmd+Shift+P) - Create tasks instantly
- ✅ **Natural language date parsing** - "tomorrow at 3pm", "in 2 hours"
- ✅ **Hashtag support** - Automatic tag extraction from #hashtags
- ✅ **Smart date detection** - Recognizes dates in task titles

### Smart Notifications
- ✅ **In-app notifications** - Banner showing overdue and upcoming tasks
- ✅ **Push notifications** - Get alerted even when app is closed
- ✅ **Smart grouping** - Multiple tasks grouped to avoid spam
- ✅ **Quiet hours** - Do-not-disturb mode for peaceful nights
- ✅ **Usage tracking** - Learns when you're most active

### Organization
- ✅ Create and share task lists
- ✅ Multiple list membership per task
- ✅ Popular tags widget
- ✅ Tag autocomplete
- ✅ Separate views for incomplete and completed tasks

### Progressive Web App
- ✅ Install on any device (desktop or mobile)
- ✅ Offline access to cached tasks
- ✅ Native app experience
- ✅ Fast loading with service worker caching

See the [Features Guide](/docs/features) for detailed information on all features.

## Getting Started

To start using AT Todo:

1. **Login with Bluesky (or your own PDS)** - Use your existing Bluesky handle (e.g., `alice.bsky.social`)
2. **Authorize the app** - Grant AT Todo permission to store tasks in your repository
3. **Start creating tasks** - Add your first task and start organizing!

See the [Getting Started guide](/docs/getting-started) for detailed instructions.

## How It Works

AT Todo uses the AT Protocol's repository system to store tasks as records. Each task is stored with the following information:

- **Title** - A brief description of the task
- **Description** - Optional detailed notes
- **Tags** - Organize tasks with flexible, user-defined tags
- **Status** - Whether the task is completed or not
- **Timestamps** - When the task was created and completed

All tasks are stored in your personal AT Protocol repository under the `app.attodo.task` lexicon.

## Using Tags

Tags help you organize and categorize your tasks. You can:

- **Add tags when creating tasks** - Enter comma-separated tags in the tags field (e.g., "work, urgent, personal")
- **Edit tags on existing tasks** - Click Edit on any task and modify the tags field
- **Filter by tag** - Click any tag to see only tasks with that tag
- **View popular tags** - See your most-used tags in the dashboard
- **Tag autocomplete** - Get suggestions from your existing tags as you type

Tags are stored in your AT Protocol repository with each task, maintaining full data ownership and portability.

## Pricing

AT Todo offers two tiers to support the community:

### Free ("Forever")
- **Price**: $0
- **Includes**:
  - Unlimited tasks
  - Unlimited lists
  - All features included
  - Your data, your control
  - No ads, ever
- **Note**: Free as long as we can keep the server up and want to develop the software. Just log in and get started!

### Supporter ⭐
- **Price**: $24 per year
- **Includes**:
  - Everything in Free
  - A gold star in the UI
  - Support development
  - Help cover server costs
  - Help keep it free for everyone
  - Our eternal gratitude
- **Status**: Coming soon! We're working on setting up supporter subscriptions.

We believe in keeping AT Todo accessible to everyone while giving those who want to support the project a way to do so.

## Privacy & Security

- Tasks are stored in your own AT Protocol repository, these are public as described in the spec
- AT Todo uses DPoP (Demonstrating Proof-of-Possession) for secure API authentication
- Your login credentials are never stored on AT Todo servers

## Support

If you encounter any issues or have questions:

- Check the [Getting Started guide](/docs/getting-started) for common questions
- Report issues on GitHub or Tangled
