<template>
  <div class="notification-bell">
    <v-menu
      v-model="showDropdown"
      :close-on-content-click="false"
      location="bottom end"
      :offset="[0, 12]"
      width="400"
      :z-index="9999"
      transition="slide-y-transition"
    >
      <template v-slot:activator="{ props }">
        <v-btn
          icon
          variant="text"
          color="white"
          class="notification-btn"
          :class="{ 'has-unread': unreadCount > 0 }"
          v-bind="props"
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
      </template>

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
          <div v-if="loading && persistentNotifications.length === 0" class="pa-4 text-center">
            <v-progress-circular indeterminate color="primary"></v-progress-circular>
            <div class="mt-2 text-body-2 text-medium-emphasis">Loading notifications...</div>
          </div>

          <div v-else-if="persistentNotifications.length === 0" class="pa-4 text-center">
            <v-icon icon="mdi-bell-off" size="48" color="grey-lighten-1"></v-icon>
            <div class="mt-2 text-body-2 text-medium-emphasis">No notifications yet</div>
          </div>

          <div v-else>
            <v-list>
              <v-list-item
                v-for="notification in displayedNotifications"
                :key="notification.id"
                :class="{ 
                  'bg-blue-lighten-5 border-s-md border-primary': notification.status === 'unread',
                  'notification-clickable': true
                }"
                @click="onNotificationClick(notification)"
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
                <v-list-item-subtitle class="notification-time">{{ formatNotificationTime(notification.created_at) }}</v-list-item-subtitle>

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

          <v-divider v-if="persistentNotifications.length > 0"></v-divider>

          <div v-if="persistentNotifications.length > 0" class="pa-3 text-center">
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
import { useNotificationStore, type Notification } from '../stores/notifications'
import { getNotificationIcon, getNotificationColor, formatNotificationTime } from '../utils/notificationUtils'

const router = useRouter()
const notificationStore = useNotificationStore()
const showDropdown = ref(false)
const dropdownLimit = 10

const {
  persistentNotifications,
  loading,
  unreadCount,
  markAsRead,
  markAllAsRead,
  loadNotifications
} = notificationStore

// Computed property to limit displayed notifications
const displayedNotifications = computed(() =>
  persistentNotifications.slice(0, dropdownLimit)
)

const onNotificationClick = (notification: Notification) => {
  if (notification.status === 'unread') {
    markAsRead(notification.id)
  }
  showDropdown.value = false
  router.push('/notifications')
}


const navigateToNotificationsPage = () => {
  showDropdown.value = false
  router.push('/notifications')
}

// Watch for dropdown state changes
watch(showDropdown, (newValue) => {
  if (newValue && persistentNotifications.length === 0) {
    loadNotifications()
  }
})

// Initial load
onMounted(() => {
  if (persistentNotifications.length === 0) {
    loadNotifications()
  }
})
</script>

<style scoped>
.notification-bell {
  position: relative;
}

.notification-btn {
  transition: all 0.2s ease;
}

.notification-btn.has-unread {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}

.notification-clickable {
  cursor: pointer;
  transition: all 0.2s ease;
}

.notification-title {
  font-weight: 500;
  line-height: 1.2;
  margin-bottom: 4px;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.notification-message {
  font-size: 0.875rem;
  line-height: 1.4;
  margin-bottom: 2px;
  word-wrap: break-word;
  overflow-wrap: break-word;
  white-space: normal;
  max-width: none;
}

.notification-time {
  font-size: 0.75rem;
  opacity: 0.7;
}
</style>