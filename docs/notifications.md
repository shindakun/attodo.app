# Notification Setup Guide

Complete guide to setting up and managing push notifications in AT Todo.

## Table of Contents

- [Getting Started](#getting-started)
- [Multi-Device Setup](#multi-device-setup)
- [Notification Settings](#notification-settings)
- [Troubleshooting](#troubleshooting)

---

## Getting Started

### Enabling Push Notifications

Push notifications allow you to receive alerts about due tasks even when AT Todo isn't open.

**Step 1: Access Settings**

1. Click the "Settings" link in the navigation bar
2. A settings dialog (modal) will open

**Step 2: Enable Notifications**

1. In the Settings dialog, scroll down to "Notification Settings"
2. You'll see your current notification status (e.g., "Not enabled")
3. Click the "Enable Push Notifications" button
4. Your browser will prompt for permission - click "Allow"

**Step 3: Configure Preferences**

Once enabled, you'll see notification preferences:

- **Timing**: Choose which types of tasks trigger notifications
  - Overdue tasks (default: on)
  - Tasks due today (default: on)
  - Tasks due within 3 days (default: off)

- **Check Frequency**: How often to check for new notifications
  - Every 15 minutes
  - Every 30 minutes (default)
  - Every hour
  - Every 2 hours

- **Quiet Hours**: Set Do Not Disturb times
  - Enable/disable quiet hours
  - Start time (default: 10 PM)
  - End time (default: 8 AM)

**Step 4: Test Your Setup**

1. Click "Send Test Notification" button
2. You should see a test notification appear
3. If successful, you'll see a confirmation toast

---

## Multi-Device Setup

AT Todo supports notifications on multiple devices - your phone, tablet, desktop, etc.

### How Multi-Device Works

- Each browser/device needs to register separately
- Notifications are sent to **all registered devices** simultaneously
- Each device maintains its own subscription
- All devices receive the same notifications

### Registering Additional Devices

**For each device you want to receive notifications:**

1. **Open AT Todo** on the new device
2. **Login** with your account
3. **Open Settings** (click the "Settings" link in the top navigation)
4. **Enable Notifications**:
   - If this is your first device: Click "Enable Push Notifications"
   - If you already have notifications enabled on another device: Click "Register This Device"
5. **Grant Permission** when your browser prompts
6. **Verify** the device appears in "Registered Devices" list

### Viewing Registered Devices

In the Settings → Notification Settings section, you'll see a "Registered Devices" list showing:

- Device count
- Browser/OS information for each device
- Registration date

Example:
```
Registered Devices:
• Device 1: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)... (added 11/22/2024)
• Device 2: Mozilla/5.0 (iPhone; CPU iPhone OS 17_0)... (added 11/22/2024)
• Device 3: Mozilla/5.0 (Windows NT 10.0; Win64; x64)... (added 11/23/2024)
```

### Re-registering a Device

If notifications stop working on a device:

1. Open Settings on that device
2. Click "Register This Device" button
3. The device will refresh its subscription
4. Test with "Send Test Notification"

### Removing Devices

To remove a device from receiving notifications:

1. **Open Settings** on any device
2. **Find the device** in the "Registered Devices" list
3. **Click the X button** next to the device you want to remove
4. **Confirm removal** when prompted
5. The device will no longer receive notifications

**Notes:**
- You can remove devices from any logged-in device (not just the device itself)
- Devices can be re-registered at any time
- Removing the last device will reset the notification UI
- If you remove the current device, you'll need to re-enable notifications

### Managing Old Devices

- Inactive device subscriptions expire automatically over time
- Failed notifications to old devices won't affect active ones
- Manual removal via the X button is the recommended cleanup method

---

## Notification Settings

### Timing Preferences

Control which tasks trigger notifications:

**Overdue Tasks** (Recommended: ON)
- Notifies when tasks are past their due date
- Highest priority - shown first
- Example: "Submit report" was due yesterday

**Tasks Due Today** (Recommended: ON)
- Notifies about tasks due within 24 hours
- Shows time until due
- Example: "Team meeting" due at 2:00 PM today

**Tasks Due Soon** (Optional)
- Notifies about tasks due within 3 days
- Helpful for planning ahead
- Example: "Project deadline" due in 2 days

**Hours Before Due** (Advanced)
- Get advance notice before tasks are due
- Range: 0-72 hours
- Example: Set to 2 hours to get notified 2 hours before due time

### Check Frequency

How often AT Todo checks for new due tasks:

- **Every 15 minutes**: Most responsive, more battery usage
- **Every 30 minutes**: Balanced (default)
- **Every hour**: Less frequent, better battery life
- **Every 2 hours**: Minimal battery impact

**Note**: Server-side checks also run every 5 minutes regardless of client settings.

### Quiet Hours (Do Not Disturb)

Prevent notifications during sleep or focus time:

**Enable Quiet Hours:**
1. Check "Enable Do Not Disturb mode"
2. Set start time (hour, 0-23)
   - Default: 22 (10 PM)
3. Set end time (hour, 0-23)
   - Default: 8 (8 AM)

**How It Works:**
- No notifications during quiet hours
- Notifications queued will appear after quiet hours end
- Works independently on each device

### Saving Changes

After configuring preferences:
1. Click "Save Preferences" button
2. You'll see a success toast
3. Settings are synced to your AT Protocol repository

---

## Troubleshooting

### Notifications Not Appearing

**Check Browser Permissions:**

1. **Chrome/Edge**:
   - Click lock icon in address bar
   - Check "Notifications" is set to "Allow"

2. **Firefox**:
   - Click lock icon → Permissions → Notifications
   - Should be "Allowed"

3. **Safari**:
   - Safari → Settings → Websites → Notifications
   - Find your domain, should be "Allow"

**Check System Settings:**

- **macOS**: System Settings → Notifications → [Your Browser]
- **Windows**: Settings → System → Notifications → [Your Browser]
- **iOS**: Settings → [Your Browser] → Notifications
- **Android**: Settings → Apps → [Your Browser] → Notifications

**Verify AT Todo Settings:**

1. Open Settings
2. Check notification status shows "Enabled ✓"
3. Verify device appears in "Registered Devices"
4. Try clicking "Register This Device" to refresh

**Test Connection:**

1. Click "Send Test Notification"
2. Check the response:
   - Success: "Test notification sent to X device(s)"
   - Failure: Error message with details
3. Check browser console for errors (F12 → Console)

### Device Not Listed

If a device doesn't appear in "Registered Devices":

1. Click "Enable Push Notifications" or "Register This Device"
2. Grant permission when prompted
3. Wait 2-3 seconds for registration
4. Refresh the Settings page
5. Device should now appear

### Notifications on Some Devices Only

If notifications work on one device but not others:

1. On the non-working device, open Settings
2. Check notification permission status
3. Click "Register This Device"
4. Send test notification
5. All devices should receive it

**Common causes:**
- Device wasn't registered (no subscription created)
- Browser permission denied
- Browser doesn't support push notifications
- System notifications disabled for browser

### Wrong Notification Times

If notifications appear at wrong times:

**Check Timezone:**
1. Verify system timezone is correct
2. Edit a task and check the time shown
3. Times should match your local timezone

**Check Task Due Times:**
1. Tasks without times trigger at midnight
2. Add specific times for better scheduling
3. Example: "tomorrow at 3pm" vs just "tomorrow"

### Too Many/Few Notifications

**Adjust Settings:**

1. **Too many**:
   - Disable "Tasks due soon" notifications
   - Increase check frequency to 2 hours
   - Enable quiet hours

2. **Too few**:
   - Enable all notification types
   - Decrease check frequency to 15 minutes
   - Check quiet hours aren't blocking

**Notification Cooldown:**
- Each task only notifies once per 12 hours
- Prevents spam for the same task
- Resets after task is completed

### Browser Compatibility Issues

**Best Support:**
- Chrome (desktop & Android)
- Edge (desktop)
- Safari (iOS 16.4+, macOS)

**Limited Support:**
- Firefox (no background sync)
- Older Safari versions
- Mobile browsers (varies)

**If using unsupported browser:**
1. Switch to Chrome or Edge for best experience
2. Or keep AT Todo open for notifications
3. Check server logs for notification sends

### Service Worker Issues

If notifications completely stop working:

**Reset Service Worker:**

1. Open browser DevTools (F12)
2. Go to Application tab → Service Workers
3. Click "Unregister" next to AT Todo worker
4. Refresh the page
5. Service worker will reinstall
6. Re-enable notifications in Settings

**Check Service Worker Status:**

1. Navigate to `/app/settings`
2. Open Console (F12)
3. Look for `[Push]` prefixed logs
4. Should see "Service worker ready"
5. Should see "Subscription registered successfully"

---

## Advanced Topics

### How Notifications Work

**Architecture:**

1. **Client-Side Check** (Browser):
   - Service worker runs periodic checks
   - Compares current time to task due dates
   - Shows notifications for matching tasks

2. **Server-Side Check** (Background Job):
   - Runs every 5 minutes
   - Checks all enabled users
   - Sends push notifications to registered devices

3. **Push Service** (Browser Vendor):
   - Chrome uses FCM (Firebase Cloud Messaging)
   - Safari uses APNs (Apple Push Notification service)
   - Delivers notifications to devices

**Data Flow:**

```
Task Due → Server Check → Push Service → Your Device → Notification
```

### Privacy & Security

**What's Stored:**

- **In Database**:
  - Your DID (user identifier)
  - Device push subscriptions (endpoints, keys)
  - Notification history (prevents spam)

- **In AT Protocol**:
  - Notification preferences
  - Task data (titles, due dates)

- **Never Stored**:
  - Notification content (generated on-demand)
  - Which notifications you viewed
  - Device location or tracking data

**Encryption:**

- Push subscriptions use public/private key encryption
- VAPID (Voluntary Application Server Identification)
- End-to-end encrypted between server and device

### Notification History

Prevent notification spam with built-in cooldown:

- Each task can only notify once per 12 hours
- Tracked in `notification_history` table
- Resets when task is completed
- Separate tracking per notification type (overdue, today, soon)

---

## Getting Help

If you're still having issues:

1. Check the [Features Guide](/docs/features) for more details
2. Review browser console for error messages
3. Check server logs for notification sending
4. Report issues on GitHub
5. Contact via Bluesky

---

## Tips for Best Experience

1. **Enable on all devices** you regularly use
2. **Set appropriate check frequency** based on urgency
3. **Use quiet hours** to avoid sleep disruption
4. **Test notifications** after first setup
5. **Keep browser updated** for best compatibility
6. **Grant persistent permissions** (don't use Incognito mode)
7. **Check "Registered Devices"** periodically to verify active devices
