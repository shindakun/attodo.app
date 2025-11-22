// Service Worker for AT Todo
const CACHE_NAME = 'attodo-v4'; // Added server ping for periodic checks
const HEALTH_CHECK_INTERVAL = 60000; // 60 seconds

// Install event - cache essential resources
self.addEventListener('install', (event) => {
  console.log('Service Worker installing...');
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll([
        '/',
        '/static/manifest.json',
        '/static/icon.svg'
      ]);
    })
  );
  self.skipWaiting();
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  console.log('Service Worker activating...');
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            console.log('Deleting old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
  self.clients.claim();
});

// Fetch event - network first, fall back to cache
self.addEventListener('fetch', (event) => {
  // For non-GET requests, just pass through to network without caching
  if (event.request.method !== 'GET') {
    event.respondWith(fetch(event.request));
    return;
  }

  // For GET requests: network first, fall back to cache
  event.respondWith(
    fetch(event.request)
      .then((response) => {
        // Cache successful GET responses
        const responseToCache = response.clone();
        caches.open(CACHE_NAME).then((cache) => {
          cache.put(event.request, responseToCache);
        });
        return response;
      })
      .catch(() => {
        // If network fails, try cache
        return caches.match(event.request);
      })
  );
});

// Health check function
async function checkHealth() {
  try {
    const response = await fetch('/health', {
      method: 'GET',
      cache: 'no-cache'
    });

    if (response.ok) {
      // Verify content type before parsing
      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        console.error('Health check returned non-JSON response:', contentType);
        throw new Error('Invalid content-type for health check');
      }

      const data = await response.json();
      console.log('Health check passed:', data);

      // Broadcast health status to all clients
      const clients = await self.clients.matchAll();
      clients.forEach(client => {
        client.postMessage({
          type: 'HEALTH_CHECK',
          status: 'healthy',
          data: data
        });
      });
    } else {
      console.warn('Health check failed with status:', response.status);

      // Broadcast unhealthy status
      const clients = await self.clients.matchAll();
      clients.forEach(client => {
        client.postMessage({
          type: 'HEALTH_CHECK',
          status: 'unhealthy',
          statusCode: response.status
        });
      });
    }
  } catch (error) {
    console.error('Health check error:', error);

    // Broadcast error status
    const clients = await self.clients.matchAll();
    clients.forEach(client => {
      client.postMessage({
        type: 'HEALTH_CHECK',
        status: 'error',
        error: error.message
      });
    });
  }
}

// Start periodic health checks when service worker activates
self.addEventListener('activate', (event) => {
  console.log('Starting health check interval...');

  // Initial health check
  checkHealth();

  // Set up periodic health checks
  setInterval(() => {
    checkHealth();
  }, HEALTH_CHECK_INTERVAL);
});

// Listen for messages from the main thread
self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'HEALTH_CHECK_NOW') {
    checkHealth();
  }
});

// ============================================================================
// NOTIFICATION SYSTEM
// ============================================================================

// Handle push notifications
self.addEventListener('push', (event) => {
  console.log('[Push] Push notification received:', event);

  // Default notification data
  let notificationData = {
    title: 'AT Todo',
    body: 'You have a new notification',
    icon: '/static/icon-192.png',
    badge: '/static/icon-192.png',
  };

  // Parse the notification payload if present
  if (event.data) {
    try {
      const payload = event.data.json();
      console.log('[Push] Payload:', payload);

      notificationData = {
        title: payload.title || notificationData.title,
        body: payload.body || notificationData.body,
        icon: payload.icon || notificationData.icon,
        badge: payload.badge || notificationData.badge,
        tag: payload.tag,
        data: payload.data,
      };
    } catch (err) {
      console.error('[Push] Failed to parse notification payload:', err);
    }
  }

  // Show the notification
  event.waitUntil(
    self.registration.showNotification(notificationData.title, {
      body: notificationData.body,
      icon: notificationData.icon,
      badge: notificationData.badge,
      tag: notificationData.tag,
      data: notificationData.data,
      vibrate: [200, 100, 200],
    })
  );
});

// Notification permission state
let notificationsEnabled = false;

// Cached settings to avoid excessive fetches
let cachedSettings = null;
let settingsCacheTime = 0;
const SETTINGS_CACHE_TTL = 5 * 60 * 1000; // 5 minutes

// Check for due tasks periodically
self.addEventListener('periodicsync', (event) => {
  if (event.tag === 'check-due-tasks') {
    event.waitUntil(
      Promise.all([
        checkDueTasksAndNotify(),  // Client-side notification display
        pingServerForCheck()        // Ping server (for future server-side push)
      ])
    );
  }
});

// Handle background sync
self.addEventListener('sync', (event) => {
  if (event.tag === 'check-tasks') {
    event.waitUntil(
      Promise.all([
        checkDueTasksAndNotify(),
        pingServerForCheck()
      ])
    );
  }
});

// Ping server for notification check (for future server-side checking)
async function pingServerForCheck() {
  try {
    await fetch('/app/push/check', {
      method: 'POST',
      credentials: 'include'
    });
  } catch (err) {
    // Silently fail - server may not be available
    console.log('[Push] Server check ping failed:', err.message);
  }
}

// Handle notification clicks
self.addEventListener('notificationclick', (event) => {
  event.notification.close();

  // Open the app
  event.waitUntil(
    clients.matchAll({ type: 'window' }).then((clientList) => {
      // If app is already open, focus it
      for (const client of clientList) {
        if (client.url.includes('/app') && 'focus' in client) {
          return client.focus();
        }
      }
      // Otherwise open a new window
      if (clients.openWindow) {
        return clients.openWindow('/app');
      }
    })
  );
});

// Get settings with caching to avoid excessive fetches
async function getSettings() {
  const now = Date.now();

  // Return cached settings if still valid
  if (cachedSettings && (now - settingsCacheTime) < SETTINGS_CACHE_TTL) {
    return cachedSettings;
  }

  // Fetch fresh settings from server
  try {
    const settingsResponse = await fetch('/app/settings', {
      credentials: 'include'
    });

    if (settingsResponse.ok) {
      cachedSettings = await settingsResponse.json();
      settingsCacheTime = now;
      return cachedSettings;
    }
  } catch (err) {
    // Network error or server unavailable
  }

  // If we have stale cached settings, use them as fallback
  if (cachedSettings) {
    return cachedSettings;
  }

  // Use default settings if no cache and fetch failed
  cachedSettings = {
    notifyOverdue: true,
    notifyToday: true,
    notifySoon: false,
    hoursBefore: 2,
    quietHoursEnabled: false,
    quietStart: 22,
    quietEnd: 8
  };
  settingsCacheTime = now;
  return cachedSettings;
}

// Check tasks and send notifications
async function checkDueTasksAndNotify() {
  try {
    // Get notification settings from AT Protocol (with caching)
    const settings = await getSettings();
    if (!settings) {
      console.warn('[Notifications] Failed to get settings');
      return; // Failed to get settings
    }

    // Check quiet hours
    if (settings.quietHoursEnabled) {
      const now = new Date();
      const hour = now.getHours();
      const quietStart = settings.quietStart || 22;
      const quietEnd = settings.quietEnd || 8;

      const isQuiet = quietStart < quietEnd
        ? (hour >= quietStart || hour < quietEnd)
        : (hour >= quietStart && hour < quietEnd);

      if (isQuiet) {
        console.log('[Notifications] Quiet hours active, skipping');
        return;
      }
    }

    // Fetch tasks as JSON
    const tasksResponse = await fetch('/app/tasks?filter=incomplete&format=json', {
      credentials: 'include',
      headers: {
        'Accept': 'application/json'
      }
    });

    if (!tasksResponse.ok) {
      console.error('[Notifications] Failed to fetch tasks:', tasksResponse.status, tasksResponse.statusText);
      return;
    }

    // Verify content type
    const contentType = tasksResponse.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
      console.error('[Notifications] Tasks endpoint returned non-JSON response:', contentType);
      const text = await tasksResponse.text();
      console.error('[Notifications] Response body (first 200 chars):', text.substring(0, 200));
      return;
    }

    const data = await tasksResponse.json();
    const tasks = Array.isArray(data) ? data : (data.tasks || []);
    console.log(`[Notifications] Fetched ${tasks.length} tasks`)

    // Group tasks by notification type
    const groups = {
      overdue: [],
      dueToday: [],
      dueSoon: []
    };

    const now = new Date();

    tasks.forEach(task => {
      if (!task.dueDate) return; // Skip tasks without due dates

      const dueDate = new Date(task.dueDate);
      const diffHours = (dueDate - now) / (1000 * 60 * 60);

      if (diffHours < 0) {
        groups.overdue.push(task);
      } else if (diffHours < 24) {
        groups.dueToday.push({
          ...task,
          dueDate,
          diffHours
        });
      } else if (diffHours < 72) {
        groups.dueSoon.push({
          ...task,
          dueDate,
          diffHours
        });
      }
    });

    // Send grouped notifications
    await sendGroupedNotifications(groups, settings);
  } catch (error) {
    console.error('[Notifications] Error checking tasks:', error);
    console.error('[Notifications] Stack trace:', error.stack);
  }
}

// Send grouped notifications to avoid spam (Phase 3.2)
async function sendGroupedNotifications(groups, settings) {
  const { overdue, dueToday, dueSoon } = groups;

  // Overdue tasks - highest priority
  if (overdue.length > 0 && settings.notifyOverdue) {
    const taskList = overdue
      .slice(0, 3) // Show up to 3 tasks
      .map(t => `• ${t.title}`)
      .join('\n');

    const moreText = overdue.length > 3 ? `\n...and ${overdue.length - 3} more` : '';

    await sendNotification(
      `${overdue.length} Overdue Task${overdue.length > 1 ? 's' : ''}`,
      taskList + moreText,
      {
        tag: 'overdue-tasks',
        badge: '/static/icon-192.png',
        renotify: true,
        requireInteraction: true // Overdue tasks are important
      }
    );
    return; // Only show one notification at a time
  }

  // Tasks due today
  if (dueToday.length > 0 && settings.notifyToday) {
    // Sort by soonest first
    dueToday.sort((a, b) => a.diffHours - b.diffHours);

    if (dueToday.length === 1) {
      // Single task - show specific time
      const task = dueToday[0];
      const hoursUntil = Math.floor(task.diffHours);
      const minutesUntil = Math.floor((task.diffHours - hoursUntil) * 60);

      let timeText = '';
      if (hoursUntil > 0) {
        timeText = `in ${hoursUntil} hour${hoursUntil > 1 ? 's' : ''}`;
      } else {
        timeText = `in ${minutesUntil} minute${minutesUntil > 1 ? 's' : ''}`;
      }

      await sendNotification(
        'Task Due Soon',
        `"${task.title}" is due ${timeText}.`,
        {
          tag: 'due-today',
          badge: '/static/icon-192.png'
        }
      );
    } else {
      // Multiple tasks - show grouped notification
      const taskList = dueToday
        .slice(0, 3)
        .map(t => {
          const hours = Math.floor(t.diffHours);
          const mins = Math.floor((t.diffHours - hours) * 60);
          const time = hours > 0 ? `${hours}h` : `${mins}m`;
          return `• ${t.title} (${time})`;
        })
        .join('\n');

      const moreText = dueToday.length > 3 ? `\n...and ${dueToday.length - 3} more` : '';

      await sendNotification(
        `${dueToday.length} Tasks Due Today`,
        taskList + moreText,
        {
          tag: 'due-today',
          badge: '/static/icon-192.png'
        }
      );
    }
    return; // Only show one notification at a time
  }

  // Tasks due soon (within 3 days)
  if (dueSoon.length > 0 && settings.notifySoon) {
    const taskList = dueSoon
      .slice(0, 3)
      .map(t => `• ${t.title}`)
      .join('\n');

    const moreText = dueSoon.length > 3 ? `\n...and ${dueSoon.length - 3} more` : '';

    await sendNotification(
      `${dueSoon.length} Task${dueSoon.length > 1 ? 's' : ''} Due Soon`,
      taskList + moreText,
      {
        tag: 'due-soon',
        badge: '/static/icon-192.png'
      }
    );
  }
}

// Helper to send notifications
async function sendNotification(title, body, options = {}) {
  const defaultOptions = {
    icon: '/static/icon-192.png',
    badge: '/static/icon-192.png',
    vibrate: [200, 100, 200],
    requireInteraction: false,
    ...options
  };

  return self.registration.showNotification(title, {
    body,
    ...defaultOptions
  });
}
