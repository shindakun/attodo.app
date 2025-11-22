# Privacy Policy

**Last Updated: November 22, 2025**

## Introduction

AT Todo ("we," "our," or "us") is committed to protecting your privacy. This Privacy Policy explains how we handle your information when you use our decentralized todo application.

## The Short Version

- **Your data stays yours.** All tasks, lists, and preferences are stored in your personal AT Protocol repository, not on our servers.
- **We don't sell your data.** We never have, and we never will.
- **Minimal data collection.** We only collect what's necessary to provide the service.
- **Open and transparent.** This policy is written in plain English.

## What is AT Protocol?

AT Todo is built on the AT Protocol (the protocol behind Bluesky), which means:

- Your data lives in **your own personal data repository (PDS)**
- Tasks and lists are stored as **public records** in your repository (as specified by the AT Protocol)
- You maintain full ownership and control of your data
- You can move your data to any AT Protocol-compatible service at any time

## Data Storage

### Data Stored in Your AT Protocol Repository

The following data is stored **in your personal AT Protocol repository**, not on AT Todo servers:

- **Tasks** - Titles, descriptions, due dates, tags, completion status
- **Lists** - List names, descriptions, task references
- **Settings** - Notification preferences, UI preferences, quiet hours settings

**Important:** Per the AT Protocol specification, this data is stored as **public records**. Anyone with your AT Protocol DID can access this data through AT Protocol APIs.

### Data Stored on AT Todo Servers

We store minimal data on our servers to provide the service:

- **Push Notification Subscriptions** - Browser push endpoints (encrypted, device-specific)
- **Supporter Status** - Whether you have an active Gold Star subscription
- **Email Address** - Only if you're a supporter, used solely to contact you about your subscription
- **Session Tokens** - Temporary tokens for authentication (HTTP-only cookies, expire automatically)

We **do not** store:
- Your password (authentication is handled by your AT Protocol provider)
- The content of your tasks or lists
- Your browsing history
- Any tracking or analytics data

## How We Use Your Information

We use your information solely to:

1. **Provide the service** - Read and write tasks/lists to your AT Protocol repository
2. **Send notifications** - Deliver push notifications about due tasks (if you enable them)
3. **Process payments** - Handle Gold Star subscriptions (via Stripe)
4. **Communicate with supporters** - Send subscription-related emails only

We **never**:
- Sell your data to third parties
- Use your data for advertising
- Track you across other websites
- Share your data except as required by law

## Third-Party Services

AT Todo uses the following third-party services:

### AT Protocol Network

- Your data is stored in AT Protocol repositories (operated by your chosen PDS provider)
- We access your repository using OAuth authentication
- See your PDS provider's privacy policy for how they handle repository data

### Stripe (Payment Processing)

- Used only for Gold Star subscriptions
- Stripe handles all payment information
- We never see or store your credit card details
- See [Stripe's Privacy Policy](https://stripe.com/privacy)

### Web Push Services

- Browser vendors (Google, Mozilla, Apple) handle push notification delivery
- We only store encrypted push subscription endpoints
- See your browser's privacy policy for push notification handling

## Cookies and Local Storage

We use:

- **Session Cookies** - HTTP-only cookies for authentication (expire when you log out)
- **Local Storage** - Browser cache for offline access to your tasks

We do **not** use:
- Tracking cookies
- Third-party advertising cookies
- Analytics cookies

## Your Rights

Because your data lives in your AT Protocol repository, you have complete control:

- **Access** - You own your data and can access it anytime via AT Protocol APIs
- **Export** - Download your data from your AT Protocol repository
- **Delete** - Delete tasks and lists directly in AT Todo or via AT Protocol
- **Move** - Migrate your data to any other AT Protocol-compatible service
- **Unsubscribe** - Disable push notifications or delete notification subscriptions anytime

To delete your AT Todo account:
1. Delete all tasks and lists in the app (or via AT Protocol APIs)
2. Disable push notifications in Settings
3. Revoke AT Todo's OAuth access in your AT Protocol provider's settings

## Data Security

We take security seriously:

- **DPoP Authentication** - Demonstrating Proof-of-Possession tokens for API requests
- **HTTPS Encryption** - All data transmission is encrypted
- **No Password Storage** - Authentication handled by your AT Protocol provider
- **Minimal Server Storage** - We don't store your tasks or personal data

## Children's Privacy

AT Todo is not directed to children under 13. We do not knowingly collect information from children under 13. If you believe we have collected information from a child under 13, please contact us immediately.

## International Users

AT Todo is operated from the United States. If you access AT Todo from outside the US, your data may be transferred to and processed in the US. By using AT Todo, you consent to this transfer.

## Changes to This Policy

We may update this Privacy Policy from time to time. We will notify you of significant changes by:
- Updating the "Last Updated" date at the top of this policy
- Posting a notice in the app (for material changes)

Continued use of AT Todo after changes constitutes acceptance of the updated policy.

## Data Retention

- **Tasks/Lists/Settings** - Stored in your AT Protocol repository indefinitely (you control deletion)
- **Push Subscriptions** - Deleted when you remove a device or disable notifications
- **Session Tokens** - Expire automatically (usually within 24 hours)
- **Supporter Data** - Retained while subscription is active, deleted upon cancellation

## Your AT Protocol Repository is Public

**Important:** The AT Protocol specification defines repository records as public by default. This means:

- Anyone who knows your DID can read your tasks and lists via AT Protocol APIs
- Tasks and lists are not encrypted in your repository
- This is a fundamental design choice of the AT Protocol
- If you need private tasks, do not store sensitive information in task titles or descriptions

## Contact Us

If you have questions about this Privacy Policy or how we handle your data:

- **Email:** [Contact via Bluesky @attodo.app](https://bsky.app/profile/attodo.app)
- **Issues:** [GitHub Issues](https://github.com/shindakun/attodo/issues) or Tangled

## Open Source

AT Todo is open source. You can review our code to see exactly how we handle your data:

- **GitHub:** [https://github.com/shindakun/attodo](https://github.com/shindakun/attodo)
- **Lexicons:** [View our data schemas](https://github.com/shindakun/attodo/tree/main/lexicons)

---

**Summary:** We store your tasks in your AT Protocol repository (public by default), keep minimal data on our servers (push subscriptions, supporter status), and never sell or track your data. You own and control everything.
