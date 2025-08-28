<template>
  <div class="notification-bell">
    <v-btn
      icon
      variant="text"
      color="white"
      class="notification-btn"
      @click="toggleDropdown"
      :class="{ 'has-unread': unreadCount > 0 }"
    >
      <v-badge
        :content="unreadCount > 99 ? '99+' : unreadCount.toString()"
        :model-value="unreadCount > 0"
        color="error"
        offset-x="8"
        offset-y="-8"
      >
        <v-icon icon="mdi-bell" size="24"></v-icon>
      </v-badge>
    </v-btn>

    <!-- Notification Dropdown -->
    <v-menu
      v-model="showDropdown"
      :close-on-content-click="false"
      location="bottom end"
      :offset="[0, 12]"
      max-width="420"
      min-width="380"
      :z-index="9999"
      transition="slide-y-transition"
    >
      <v-card class="notification-dropdown">
        <v-card-title class="d-flex align-center justify-space-between pa-4 bg-primary text-white">
          <span class="text-h6 font-weight-medium">Notifications</span>
          <v-btn
            v-if="unreadCount > 0"
            size="small"
            variant="text"
            color="white"
            @click="markAllAsRead"
            :loading="loading"
            class="text-caption"
          >
            Mark all read
          </v-btn>
        </v-card-title>

        <v-divider></v-divider>

        <div class="notification-list">
          <div v-if="loading && notifications.length === 0" class="pa-4 text-center">
            <v-progress-circular indeterminate color="primary"></v-progress-circular>
            <div class="mt-2 text-body-2 text-medium-emphasis">Loading notifications...</div>
          </div>

          <div v-else-if="notifications.length === 0" class="pa-4 text-center">
            <v-icon icon="mdi-bell-off" size="48" color="grey-lighten-1"></v-icon>
            <div class="mt-2 text-body-2 text-medium-emphasis">No notifications yet</div>
          </div>

          <div v-else class="notification-items">
            <v-list>
              <v-list-item
                v-for="notification in notifications"
                :key="notification.id"
                :class="{ 'bg-blue-lighten-5 border-s-md border-primary': notification.status === 'unread' }"
                @click="handleNotificationClick(notification)"
                class="py-2"
                :ripple="true"
              >
                <template v-slot:prepend>
                  <v-avatar size="40" :color="getNotificationColor(notification.type)" class="notification-icon">
                    <v-icon :icon="getNotificationIcon(notification.type)" color="white"></v-icon>
                  </v-avatar>
                </template>

                <v-list-item-title class="notification-title">{{ notification.title }}</v-list-item-title>
                <v-list-item-subtitle class="notification-message">{{ notification.message }}</v-list-item-subtitle>
                <v-list-item-subtitle class="notification-time">{{ formatTime(notification.created_at) }}</v-list-item-subtitle>

                <template v-slot:append>
                  <v-btn
                    v-if="notification.status === 'unread'"
                    icon
                    size="small"
                    variant="text"
                    @click.stop="markAsRead(notification.id)"
                    color="success"
                  >
                    <v-icon icon="mdi-check" size="16"></v-icon>
                  </v-btn>
                </template>
              </v-list-item>
            </v-list>
          </div>

          <v-divider v-if="notifications.length > 0"></v-divider>

          <div v-if="notifications.length > 0" class="pa-3 text-center">
            <v-btn
              variant="text"
              color="primary"
              size="small"
              @click="navigateToNotificationsPage"
              prepend-icon="mdi-eye"
            >
              See All Notifications
            </v-btn>
          </div>
        </div>
      </v-card>
    </v-menu>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { usePersistentNotificationStore, type PersistentNotification } from '../stores/persistentNotifications'
import { formatDistanceToNow } from 'date-fns'

const router = useRouter()
const notificationStore = usePersistentNotificationStore()
const showDropdown = ref(false)
const currentPage = ref(0)
const pageSize = 20

const {
  notifications,
  loading,
  unreadCount,
  fetchNotifications,
  markAsRead,
  markAllAsRead,
  clearAll
} = notificationStore

const toggleDropdown = () => {
  console.log('Toggle dropdown clicked, current state:', showDropdown.value)
  showDropdown.value = !showDropdown.value
  if (showDropdown.value && notifications.length === 0) {
    console.log('Fetching notifications...')
    fetchNotifications(pageSize, 0)
  }
}

const handleNotificationClick = (notification: PersistentNotification) => {
  if (notification.status === 'unread') {
    markAsRead(notification.id)
  }

  // Handle notification action based on type
  if (notification.task_id) {
    // Navigate to task details or handle task-specific action
    console.log('Navigate to task:', notification.task_id)
  }
}

const getNotificationIcon = (type: string) => {
  switch (type) {
    case 'deployment_update': return 'mdi-rocket-launch'
    case 'task_update': return 'mdi-cog'
    case 'connected': return 'mdi-link'
    case 'error': return 'mdi-alert-circle'
    default: return 'mdi-bell'
  }
}

const getNotificationColor = (type: string) => {
  switch (type) {
    case 'deployment_update': return 'success'
    case 'task_update': return 'info'
    case 'connected': return 'primary'
    case 'error': return 'error'
    default: return 'grey'
  }
}

const formatTime = (timestamp: string) => {
  try {
    return formatDistanceToNow(new Date(timestamp), { addSuffix: true })
  } catch {
    return 'Unknown time'
  }
}

const loadMore = () => {
  currentPage.value++
  fetchNotifications(pageSize, currentPage.value * pageSize)
}

const navigateToNotificationsPage = () => {
  showDropdown.value = false
  router.push('/notifications')
}

// Watch for dropdown state changes
watch(showDropdown, (newValue) => {
  if (newValue && notifications.length === 0) {
    fetchNotifications(pageSize, 0)
  }
})

// Initial load
onMounted(() => {
  console.log('NotificationBell mounted, unreadCount:', unreadCount)
  // Always fetch notifications on mount to get the current count
  fetchNotifications(pageSize, 0)
})
</script>

<style scoped>
.notification-btn.has-unread {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}
</style>
