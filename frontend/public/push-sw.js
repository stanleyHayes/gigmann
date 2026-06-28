/* eslint-disable no-undef */
// Web Push handlers (GEC-69). Imported into the generated Workbox service worker
// via vite-plugin-pwa's `workbox.importScripts`, so the existing offline/caching
// config is untouched. The API only ever sends severity=critical alerts here
// ("quiet by default"), with figures coming straight from the deterministic
// alert payload — the notification never invents a number.

self.addEventListener('push', (event) => {
  let payload = {}
  try {
    payload = event.data ? event.data.json() : {}
  } catch {
    payload = {}
  }
  const title = payload.title || 'Gigmann — critical alert'
  const options = {
    body: payload.body || '',
    tag: payload.alertId || 'gigmann-alert',
    data: { url: payload.url || '/alerts' },
    icon: '/pwa-192x192.png',
    badge: '/pwa-192x192.png',
    requireInteraction: true,
  }
  event.waitUntil(self.registration.showNotification(title, options))
})

self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  const url = (event.notification.data && event.notification.data.url) || '/alerts'
  event.waitUntil(
    self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      for (const client of clientList) {
        if ('focus' in client) {
          if ('navigate' in client) client.navigate(url)
          return client.focus()
        }
      }
      return self.clients.openWindow ? self.clients.openWindow(url) : undefined
    }),
  )
})
