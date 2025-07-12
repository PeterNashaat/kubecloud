import { ref, onMounted, onUnmounted } from 'vue'
import { useNotificationStore } from '../stores/notifications'
import { useUserStore } from '../stores/user'

export function useDeploymentEvents() {
  const eventSource = ref<EventSource | null>(null)
  const notificationStore = useNotificationStore()

  function connect() {
    if (eventSource.value) return
    // Use backend base URL from environment variable or fallback
    const backendBaseUrl = import.meta.env.VITE_BACKEND_URL || 'http://localhost:8080'
    const userStore = useUserStore()
    const token = userStore.token || ''
    const url = backendBaseUrl + '/api/v1/events?token=' + encodeURIComponent(token)
    console.log('Connecting to EventSource at', url)
    eventSource.value = new EventSource(url, { withCredentials: true })

    eventSource.value.onopen = () => {
      console.log('EventSource connection opened')
    }

    eventSource.value.onmessage = (event) => {
      console.log('EventSource message received:', event.data)
      try {
        const data = JSON.parse(event.data)
        const type = data.type || 'info'
        const message = data.message || JSON.stringify(data)
        if (type === 'success') {
          notificationStore.success('Deployment', message)
        } else if (type === 'error') {
          notificationStore.error('Deployment Error', message)
        } else {
          notificationStore.info('Deployment', message)
        }
      } catch (err) {
        notificationStore.info('Deployment', event.data)
      }
    }
    eventSource.value.onerror = (err) => {
      disconnect()
    }
  }

  function disconnect() {
    if (eventSource.value) {
      eventSource.value.close()
      eventSource.value = null
    }
  }

  onMounted(connect)
  onUnmounted(disconnect)

  return { connect, disconnect }
}
