import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../utils/api'

export interface PersistentNotification {
  id: number
  type: 'deployment_update' | 'task_update' | 'connected' | 'error'
  title: string
  message: string
  data?: string
  task_id?: string
  status: 'read' | 'unread'
  created_at: string
  read_at?: string
}

export const usePersistentNotificationStore = defineStore('persistentNotifications', () => {
  const notifications = ref<PersistentNotification[]>([])
  const loading = ref(false)
  const unreadCount = computed(() => notifications.value.filter(n => n.status === 'unread').length)

  // Fetch all notifications
  const fetchNotifications = async (limit = 20, offset = 0) => {
    try {
      loading.value = true
      const response = await api.get<{data: {notifications: PersistentNotification[], limit: number, offset: number, count: number}}>(`/v1/notifications?limit=${limit}&offset=${offset}`, {
        requiresAuth: true
      })
      
      console.log('Notification response:', response)
      
      if (response.status === 200 && response.data?.data?.notifications) {
        notifications.value = response.data.data.notifications
        console.log('Notifications loaded:', notifications.value)
      } else {
        console.log('No notifications in response or invalid status')
        console.log('Response data structure:', response.data)
      }
    } catch (error) {
      console.error('Failed to fetch notifications:', error)
    } finally {
      loading.value = false
    }
  }

  // Fetch unread notifications
  const fetchUnreadNotifications = async (limit = 20, offset = 0) => {
    try {
      loading.value = true
      const response = await api.get<{data: {notifications: PersistentNotification[], limit: number, offset: number, count: number}}>(`/v1/notifications/unread?limit=${limit}&offset=${offset}`, {
        requiresAuth: true
      })
      
      if (response.status === 200 && response.data?.data?.notifications) {
        // Merge with existing notifications, avoiding duplicates
        const newNotifications = response.data.data.notifications
        const existingIds = new Set(notifications.value.map(n => n.id))
        
        newNotifications.forEach((notification: PersistentNotification) => {
          if (!existingIds.has(notification.id)) {
            notifications.value.unshift(notification)
          }
        })
      }
    } catch (error) {
      console.error('Failed to fetch unread notifications:', error)
    } finally {
      loading.value = false
    }
  }

  // Mark notification as read
  const markAsRead = async (notificationId: number) => {
    try {
      const response = await api.put(`/v1/notifications/${notificationId}/read`, undefined, {
        requiresAuth: true
      })
      
      if (response.status === 200) {
        const notification = notifications.value.find(n => n.id === notificationId)
        if (notification) {
          notification.status = 'read'
          notification.read_at = new Date().toISOString()
        }
      }
    } catch (error) {
      console.error('Failed to mark notification as read:', error)
    }
  }

  // Mark all notifications as read
  const markAllAsRead = async () => {
    try {
      const response = await api.put('/v1/notifications/read-all', undefined, {
        requiresAuth: true
      })
      
      if (response.status === 200) {
        notifications.value.forEach(notification => {
          notification.status = 'read'
          notification.read_at = new Date().toISOString()
        })
      }
    } catch (error) {
      console.error('Failed to mark all notifications as read:', error)
    }
  }

  // Delete notification
  const deleteNotification = async (notificationId: number) => {
    try {
      const response = await api.delete(`/v1/notifications/${notificationId}`, {
        requiresAuth: true
      })
      
      if (response.status === 200) {
        const index = notifications.value.findIndex(n => n.id === notificationId)
        if (index > -1) {
          notifications.value.splice(index, 1)
        }
      }
    } catch (error) {
      console.error('Failed to delete notification:', error)
    }
  }

  // Add new notification (for real-time updates)
  const addNotification = (notification: PersistentNotification) => {
    // Check if notification already exists
    const existingIndex = notifications.value.findIndex(n => n.id === notification.id)
    if (existingIndex > -1) {
      notifications.value[existingIndex] = notification
    } else {
      notifications.value.unshift(notification)
    }
  }

  // Clear all notifications
  const clearAll = async () => {
    try {
      const response = await api.put('/v1/notifications', undefined, {
        requiresAuth: true
      })
      
      if (response.status === 200) {
        notifications.value = []
      }
    } catch (error) {
      console.error('Failed to clear all notifications:', error)
    }
  }

  return {
    notifications,
    loading,
    unreadCount,
    fetchNotifications,
    fetchUnreadNotifications,
    markAsRead,
    markAllAsRead,
    deleteNotification,
    addNotification,
    clearAll
  }
})
