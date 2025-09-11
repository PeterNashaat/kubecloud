import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../utils/api'

// Backend notification types
export type NotificationType = 'deployment' | 'billing' | 'user' | 'connected'
export type NotificationSeverity = 'info' | 'error' | 'warning' | 'success'
export type NotificationStatus = 'read' | 'unread'

// Backend notification response interface
export interface BackendNotification {
  id: string
  task_id?: string
  type: NotificationType
  severity: NotificationSeverity
  payload: Record<string, string>
  status: NotificationStatus
  created_at: string
  read_at?: string
}

// Unified notification interface for frontend
export interface Notification {
  id: string
  type: NotificationType
  severity: NotificationSeverity
  payload: Record<string, string>
  status: NotificationStatus
  created_at: string
  read_at?: string
  task_id?: string
  // Frontend-specific fields
  duration?: number
  persistent?: boolean
}

export const useNotificationStore = defineStore('notifications', () => {
  // State
  const notifications = ref<Notification[]>([])
  const loading = ref(false)
  const activeTimeouts = ref<Map<string, NodeJS.Timeout>>(new Map())
  
  // Computed
  const unreadCount = computed(() => {
    return notifications.value.filter(n => n.persistent && n.status === 'unread').length
  })
  
  const toastNotifications = computed(() => 
    notifications.value.filter(n => !n.persistent)
  )
  
  const persistentNotifications = computed(() => 
    notifications.value.filter(n => n.persistent)
  )

  // Helper function to convert backend notification to frontend format
  const convertBackendNotification = (backendNotif: BackendNotification): Notification => ({
    ...backendNotif,
    persistent: true
  })

  // Note: SSE toasts use the convenience methods below; no generic payload helper needed

  // Core functions
  const addNotification = (notification: Omit<Notification, 'id'>) => {
    const id = notification.persistent ? `temp-${Date.now()}` : `toast-${Date.now()}-${Math.random()}`
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

  const markAsRead = async (id: string) => {
    const notification = notifications.value.find(n => n.id === id)
    if (notification && notification.persistent && notification.status === 'unread') {
      try {
        await api.patch(`/v1/notifications/${id}/read`, undefined, { requiresAuth: true })
        notification.status = 'read'
        notification.read_at = new Date().toISOString()
      } catch (error) {
        console.error('Failed to mark notification as read:', error)
        throw error
      }
    }
  }

  const markAsUnread = async (id: string) => {
    const notification = notifications.value.find(n => n.id === id)
    if (notification && notification.persistent && notification.status === 'read') {
      try {
        await api.patch(`/v1/notifications/${id}/unread`, undefined, { requiresAuth: true })
        notification.status = 'unread'
        notification.read_at = undefined
      } catch (error) {
        console.error('Failed to mark notification as unread:', error)
        throw error
      }
    }
  }

  const markAllAsRead = async () => {
    const unreadNotifications = notifications.value.filter(n => n.persistent && n.status === 'unread')
    if (unreadNotifications.length === 0) return

    try {
      await api.patch('/v1/notifications/read-all', undefined, { requiresAuth: true })
      unreadNotifications.forEach(notification => {
        notification.status = 'read'
        notification.read_at = new Date().toISOString()
      })
    } catch (error) {
      console.error('Failed to mark all notifications as read:', error)
      throw error
    }
  }

  const clearAll = async () => {
    const persistentNotifications = notifications.value.filter(n => n.persistent)
    if (persistentNotifications.length === 0) return

    try {
      await api.delete('/v1/notifications', { requiresAuth: true })
      notifications.value = notifications.value.filter(n => !n.persistent)
    } catch (error) {
      console.error('Failed to clear all notifications:', error)
    }
  }

  const deleteNotification = async (id: string) => {
    const notification = notifications.value.find(n => n.id === id)
    if (notification && notification.persistent) {
      try {
        await api.delete(`/v1/notifications/${id}`, { requiresAuth: true })
        removeNotification(id)
      } catch (error) {
        console.error('Failed to delete notification:', error)
        throw error
      }
    }
  }

  // Load persistent notifications from server
  const loadNotifications = async () => {
    if (loading.value) return
    
    try {
      loading.value = true
      const response = await api.get('/v1/notifications?limit=50', { requiresAuth: true })
      
      if (response.status === 200 && (response.data as any)?.data?.notifications) {
        const serverNotifications = (response.data as any).data.notifications.map((n: BackendNotification) => 
          convertBackendNotification(n)
        )
        
        // Replace persistent notifications, keep toast notifications
        const toastNotifications = notifications.value.filter(n => !n.persistent)
        notifications.value = [
          ...serverNotifications,
          ...toastNotifications
        ]
      }
    } catch (error) {
      console.error('Failed to load notifications:', error)
    } finally {
      loading.value = false
    }
  }

  // Load unread notifications from server
  const loadUnreadNotifications = async () => {
    if (loading.value) return
    
    try {
      loading.value = true
      const response = await api.get('/v1/notifications/unread?limit=50', { requiresAuth: true })
      
      if (response.status === 200 && (response.data as any)?.data?.notifications) {
        const serverNotifications = (response.data as any).data.notifications.map((n: BackendNotification) => 
          convertBackendNotification(n)
        )
        
        // Replace persistent notifications, keep toast notifications
        const toastNotifications = notifications.value.filter(n => !n.persistent)
        notifications.value = [
          ...serverNotifications,
          ...toastNotifications
        ]
      }
    } catch (error) {
      console.error('Failed to load unread notifications:', error)
    } finally {
      loading.value = false
    }
  }

  // Convenience methods for toast notifications
  const success = (title: string, message: string) => 
    addNotification({ 
      type: 'user', 
      severity: 'success', 
      payload: { title, message }, 
      status: 'read',
      persistent: false, 
      created_at: new Date().toISOString() 
    })
  
  const error = (title: string, message: string) => 
    addNotification({ 
      type: 'user', 
      severity: 'error', 
      payload: { title, message }, 
      status: 'read',
      persistent: false, 
      created_at: new Date().toISOString() 
    })
  
  const warning = (title: string, message: string) => 
    addNotification({ 
      type: 'user', 
      severity: 'warning', 
      payload: { title, message }, 
      status: 'read',
      persistent: false, 
      created_at: new Date().toISOString() 
    })
  
  const info = (title: string, message: string) => 
    addNotification({ 
      type: 'user', 
      severity: 'info', 
      payload: { title, message }, 
      status: 'read',
      persistent: false, 
      created_at: new Date().toISOString() 
    })

  // Cleanup function to prevent memory leaks
  const cleanup = () => {
    activeTimeouts.value.forEach(timeout => clearTimeout(timeout))
    activeTimeouts.value.clear()
  }

  // Reset store state (used on logout)
  const reset = () => {
    cleanup()
    notifications.value = []
    loading.value = false
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
    markAsUnread,
    markAllAsRead,
    clearAll,
    deleteNotification,
    loadNotifications,
    loadUnreadNotifications,
    
    // Helper functions
    convertBackendNotification,
    
    // Toast convenience methods
    success,
    error,
    warning,
    info,
    
    // Cleanup
    cleanup,
    reset
  }
})