import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useNotificationStore } from '../stores/notifications'
import { usePersistentNotificationStore } from '../stores/persistentNotifications'
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
  const uiToast = useNotificationStore()
  const persistentStore = usePersistentNotificationStore()
  const userStore = useUserStore()
  const clusterStore = useClusterStore()
  const { fetchRentedNodes } = useNodeManagement()

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

    // Map known event kinds to UI toasts and/or persistent notifications
    if (type === 'workflow_update') {
      const message = evt.message || evt.data?.message
      if (message) {
        const text = message.toLowerCase()
        if (text.includes('failed')) {
          uiToast.error('Workflow', message)
        } else if (text.includes('completed')) {
          uiToast.success('Workflow', message)
        } else {
          uiToast.info('Workflow', message)
        }
      }

      refreshClusterData()
      return
    }

    // Default channel: create a persistent notification entry so bell shows it
    persistentStore.addNotification({
      id: Date.now(), // temporary client id; server-provided ids will overwrite on fetch
      type: (evt.type as any) || 'task_update',
      title: evt.data?.status || 'Notification',
      message: evt.data?.message || evt.message || 'New notification',
      data: JSON.stringify(evt.data || {}),
      task_id: evt.taskId || '',
      status: 'unread',
      created_at: new Date().toISOString()
    } as any)
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
