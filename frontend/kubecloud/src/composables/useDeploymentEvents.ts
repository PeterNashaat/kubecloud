import { ref, onMounted, onUnmounted } from 'vue'
import { useNotificationStore } from '../stores/notifications'

export function useDeploymentEvents() {
  const eventSource = ref<EventSource | null>(null)
  const notificationStore = useNotificationStore()

  function connect() {
    if (eventSource.value) return
    eventSource.value = new EventSource('/api/v1/events', { withCredentials: true })
    eventSource.value.onmessage = (event) => {
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
    eventSource.value.onerror = () => {
      disconnect()
      setTimeout(connect, 3000)
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