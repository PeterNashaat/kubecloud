import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useNotificationStore } from '../stores/notifications'
import { useUserStore } from '../stores/user'
import { useClusterStore } from '../stores/clusters'
import { useNodeManagement } from './useNodeManagement'

export interface SseEvent {
  type: string
  data: any
  message?: string
  timestamp?: string
  taskId?: string
}

export function useSseEvents() {
  const eventSource = ref<EventSource | null>(null)
  const notificationStore = useNotificationStore()
  const userStore = useUserStore()
  const clusterStore = useClusterStore()

  const isConnected = ref(false)
  const reconnectAttempts = ref(0)
  const maxReconnectAttempts = 5
  const reconnectDelayMs = 2000

  function connect() {
    if (eventSource.value || isConnected.value) return

    const backendBaseUrl = (typeof window !== 'undefined' && (window as any).__ENV__?.VITE_API_BASE_URL) || import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'
    const token = userStore.token || localStorage.getItem('token') || ''
    if (!token) return

    const url = `${backendBaseUrl}/v1/events?token=${encodeURIComponent(token)}`
    eventSource.value = new EventSource(url, { withCredentials: true })

    eventSource.value.onopen = () => {
      isConnected.value = true
      reconnectAttempts.value = 0
    }

    eventSource.value.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data) as SseEvent
        handleEvent(payload)
      } catch (e) {
        console.error('Failed to parse SSE message', e)
      }
    }

    eventSource.value.onerror = (err) => {
      console.error('SSE connection error:', err)
      isConnected.value = false

      if (reconnectAttempts.value < maxReconnectAttempts) {
        setTimeout(() => {
          reconnectAttempts.value++
          disconnect()
          connect()
        }, reconnectDelayMs * Math.max(1, reconnectAttempts.value))
      }
    }
  }

  function handleEvent(evt: SseEvent) {
    const type = evt.type || 'info'

    // Skip noise
    if (type === 'connected') {
      isConnected.value = true
      return
    }

    // Map known event kinds to UI toasts only (SSE should not persist)
    if (type === 'workflow_update') {
      const message = evt.message || evt.data?.message
      if (message) {
        const text = message.toLowerCase()
        if (text.includes('failed')) {
          notificationStore.error('Workflow', message)
        } else if (text.includes('completed')) {
          notificationStore.success('Workflow', message)
        } else {
          notificationStore.info('Workflow', message)
        }
      }

      refreshClusterData()
      return
    }

    // Default channel: show as toast only (non-persistent)
    const title = evt.data?.status || 'Notification'
    const message = evt.data?.message || evt.message || 'New notification'
    const lower = `${title} ${message}`.toLowerCase()
    if (lower.includes('fail')) {
      notificationStore.error(title, message)
    } else if (lower.includes('complete') || lower.includes('success')) {
      notificationStore.success(title, message)
    } else if (lower.includes('warn')) {
      notificationStore.warning(title, message)
    } else {
      notificationStore.info(title, message)
    }
  }

  async function refreshClusterData() {
    await clusterStore.fetchClusters()
  }

  function disconnect() {
    if (eventSource.value) {
      eventSource.value.close()
      eventSource.value = null
    }
    isConnected.value = false
  }

  // Auto manage lifecycle with token
  watch(() => userStore.token, (newToken) => {
    if (newToken && !isConnected.value) {
      connect()
    }
    if (!newToken && isConnected.value) {
      disconnect()
    }
  }, { immediate: true })

  onMounted(() => {
    setTimeout(() => {
      if (userStore.token && !isConnected.value) connect()
    }, 100)
  })

  onUnmounted(() => {
    disconnect()
  })

  return { connect, disconnect, isConnected }
}
