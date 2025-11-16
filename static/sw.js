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
