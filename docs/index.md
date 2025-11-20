# Documentation

Welcome to AT Todo documentation! AT Todo is a decentralized todo application built on the AT Protocol, allowing you to own and control your task data.

## What is AT Todo?

AT Todo leverages the AT Protocol (the protocol behind Bluesky) to store your tasks in your own personal data repository. This means:

- **You own your data** - Tasks are stored in your AT Protocol repository
- **Decentralized** - No central server owns your information
- **Portable** - Your data can be accessed by any AT Protocol-compatible application

## Features

- ✅ Create and manage tasks with titles and descriptions
- ✅ Mark tasks as complete/incomplete
- ✅ Edit existing tasks
- ✅ Delete tasks
- ✅ Organize tasks with tags
- ✅ Filter tasks by tags
- ✅ Create and share task lists
- ✅ Separate views for incomplete and completed tasks
- ✅ Progressive Web App (PWA) support - install on your device

## Getting Started

To start using AT Todo:

1. **Login with Bluesky** - Use your existing Bluesky handle (e.g., `alice.bsky.social`)
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

## Privacy & Security

- Tasks are stored in your own AT Protocol repository
- AT Todo uses DPoP (Demonstrating Proof-of-Possession) for secure API authentication
- Your login credentials are never stored on AT Todo servers

## Support

If you encounter any issues or have questions:

- Check the [Getting Started guide](/docs/getting-started) for common questions
- Report issues on GitHub or Tangled

## About

AT Todo is a demonstration of building applications on the AT Protocol. It showcases:

- OAuth authentication with AT Protocol
- DPoP-secured API requests
- Progressive Web App capabilities
- Decentralized data storage

Made with ❤ in PDX!
