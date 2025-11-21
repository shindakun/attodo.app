// Service Worker for AT Todo
const CACHE_NAME = 'attodo-v1';
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
  event.respondWith(
    fetch(event.request)
      .then((response) => {
        // Clone the response before caching
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

// Notification permission state
let notificationsEnabled = false;

// Check for due tasks periodically
self.addEventListener('periodicsync', (event) => {
  if (event.tag === 'check-due-tasks') {
    event.waitUntil(checkDueTasksAndNotify());
  }
});

// Handle background sync
self.addEventListener('sync', (event) => {
  if (event.tag === 'check-tasks') {
    event.waitUntil(checkDueTasksAndNotify());
  }
});

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

// Check tasks and send notifications
async function checkDueTasksAndNotify() {
  try {
    // Get notification settings from AT Protocol
    let settings = null;
    try {
      const settingsResponse = await fetch('/app/settings', {
        credentials: 'include'
      });
      if (settingsResponse.ok) {
        settings = await settingsResponse.json();
      }
    } catch (err) {
      // Could not load settings, using defaults
    }

    // Use default settings if none available
    if (!settings) {
      settings = {
        notifyOverdue: true,
        notifyToday: true,
        notifySoon: false,
        hoursBefore: 2,
        quietHoursEnabled: false
      };
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
        return;
      }
    }

    // Fetch tasks as JSON
    const tasksResponse = await fetch('/app/tasks?filter=incomplete&format=json', {
      credentials: 'include'
    });

    if (!tasksResponse.ok) {
      return;
    }

    const data = await tasksResponse.json();
    const tasks = data.tasks || data; // Handle both {tasks: []} and [] formats

    const now = new Date();
    let overdueCount = 0;
    let dueTodayCount = 0;
    let nextDueTasks = [];

    tasks.forEach(task => {
      if (!task.dueDate) return; // Skip tasks without due dates

      const dueDate = new Date(task.dueDate);
      const diffHours = (dueDate - now) / (1000 * 60 * 60);

      if (diffHours < 0) {
        overdueCount++;
      } else if (diffHours < 24) {
        dueTodayCount++;
        nextDueTasks.push({
          title: task.title,
          dueDate,
          diffHours
        });
      }
    });

    // Send notifications based on what we found and user preferences
    if (overdueCount > 0 && settings.notifyOverdue) {
      await sendNotification(
        'Overdue Tasks!',
        `You have ${overdueCount} overdue task${overdueCount > 1 ? 's' : ''}.`,
        { tag: 'overdue-tasks', badge: '/static/icon-192.png' }
      );
    } else if (dueTodayCount > 0 && settings.notifyToday) {
      // Sort by soonest first
      nextDueTasks.sort((a, b) => a.diffHours - b.diffHours);
      const soonest = nextDueTasks[0];

      const hoursUntil = Math.floor(soonest.diffHours);
      const minutesUntil = Math.floor((soonest.diffHours - hoursUntil) * 60);

      let timeText = '';
      if (hoursUntil > 0) {
        timeText = `in ${hoursUntil} hour${hoursUntil > 1 ? 's' : ''}`;
      } else {
        timeText = `in ${minutesUntil} minute${minutesUntil > 1 ? 's' : ''}`;
      }

      await sendNotification(
        'Task Due Soon',
        `"${soonest.title}" is due ${timeText}.`,
        { tag: 'due-soon', badge: '/static/icon-192.png' }
      );
    }
  } catch (error) {
    // Error checking tasks for notifications
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
