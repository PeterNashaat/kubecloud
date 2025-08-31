import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../utils/api'

// Unified notification interface
export interface Notification {
  id: number | string
  type: 'success' | 'error' | 'warning' | 'info' | 'deployment_update' | 'task_update' | 'connected'
  title: string
  message: string
  status?: 'read' | 'unread'
  created_at: string
  read_at?: string
  duration?: number
  persistent?: boolean // true for bell notifications, false for toast notifications
}

export const useNotificationStore = defineStore('notifications', () => {
  // State
  const notifications = ref<Notification[]>([])
  const loading = ref(false)
  const activeTimeouts = ref<Map<string, NodeJS.Timeout>>(new Map())
  
  // Computed
  const unreadCount = computed(() => {
    const count = notifications.value.filter(n => n.persistent && n.status === 'unread').length
    return count
  })
  
  const toastNotifications = computed(() => 
    notifications.value.filter(n => !n.persistent)
  )
  
  const persistentNotifications = computed(() => 
    notifications.value.filter(n => n.persistent)
  )

  // Core functions
  const addNotification = (notification: Omit<Notification, 'id'>) => {
    const id = notification.persistent ? Date.now() : `${Date.now()}-${Math.random()}`
    const newNotification: Notification = {
      ...notification,
      id,
      persistent: notification.persistent || false,
      duration: notification.duration || (notification.persistent ? undefined : 5000)
    }
    
    notifications.value.unshift(newNotification)
    
    // Auto-remove toast notifications
    if (!notification.persistent && newNotification.duration) {
      const timeout = setTimeout(() => {
        removeNotification(String(id))
        activeTimeouts.value.delete(String(id))
      }, newNotification.duration)
      activeTimeouts.value.set(String(id), timeout)
    }
    
    return id
  }

  const removeNotification = (id: number | string) => {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
      // Clear timeout if exists
      const timeout = activeTimeouts.value.get(String(id))
      if (timeout) {
        clearTimeout(timeout)
        activeTimeouts.value.delete(String(id))
      }
      notifications.value.splice(index, 1)
    }
  }

  const markAsRead = async (id: number | string) => {
    const notification = notifications.value.find(n => n.id === id)
    if (notification && notification.persistent && notification.status === 'unread') {
      try {
        await api.put(`/v1/notifications/${id}/read`, undefined, { requiresAuth: true })
        notification.status = 'read'
        notification.read_at = new Date().toISOString()
      } catch (error) {
        console.error('Failed to mark notification as read:', error)
      }
    }
  }

  const markAllAsRead = async () => {
    const unreadNotifications = notifications.value.filter(n => n.persistent && n.status === 'unread')
    if (unreadNotifications.length === 0) return

    try {
      await api.put('/v1/notifications/read-all', undefined, { requiresAuth: true })
      unreadNotifications.forEach(notification => {
        notification.status = 'read'
        notification.read_at = new Date().toISOString()
      })
    } catch (error) {
      console.error('Failed to mark all notifications as read:', error)
    }
  }

  const clearAll = async () => {
    const persistentNotifications = notifications.value.filter(n => n.persistent)
    if (persistentNotifications.length === 0) return

    try {
      await api.put('/v1/notifications', undefined, { requiresAuth: true })
      notifications.value = notifications.value.filter(n => !n.persistent)
    } catch (error) {
      console.error('Failed to clear all notifications:', error)
    }
  }

  // Load persistent notifications from server
  const loadNotifications = async () => {
    if (loading.value) return
    
    try {
      loading.value = true
      const response = await api.get('/v1/notifications?limit=50', { requiresAuth: true })
      
      if (response.status === 200 && (response.data as any)?.data?.notifications) {
        const serverNotifications = (response.data as any).data.notifications.map((n: any) => ({
          ...n,
          persistent: true
        }))
        
        // Replace persistent notifications, keep toast notifications
        notifications.value = [
          ...serverNotifications,
          ...notifications.value.filter(n => !n.persistent)
        ]
      }
    } catch (error) {
      console.error('Failed to load notifications:', error)
    } finally {
      loading.value = false
    }
  }

  // Convenience methods for toast notifications
  const success = (title: string, message: string) => 
    addNotification({ type: 'success', title, message, persistent: false, created_at: new Date().toISOString() })
  
  const error = (title: string, message: string) => 
    addNotification({ type: 'error', title, message, persistent: false, created_at: new Date().toISOString() })
  
  const warning = (title: string, message: string) => 
    addNotification({ type: 'warning', title, message, persistent: false, created_at: new Date().toISOString() })
  
  const info = (title: string, message: string) => 
    addNotification({ type: 'info', title, message, persistent: false, created_at: new Date().toISOString() })

  // Cleanup function to prevent memory leaks
  const cleanup = () => {
    activeTimeouts.value.forEach(timeout => clearTimeout(timeout))
    activeTimeouts.value.clear()
  }

  return {
    // State
    notifications,
    loading,
    unreadCount,
    toastNotifications,
    persistentNotifications,
    
    // Actions
    addNotification,
    removeNotification,
    markAsRead,
    markAllAsRead,
    clearAll,
    loadNotifications,
    
    // Toast convenience methods
    success,
    error,
    warning,
    info,
    
    // Cleanup
    cleanup
  }
})