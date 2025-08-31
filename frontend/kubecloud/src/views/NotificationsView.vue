<template>
  <div class="notifications-page">
    <div class="container mx-auto pa-6" style="margin-top: 6rem;">
      <!-- Header -->
      <div class="d-flex align-center justify-space-between mb-6">
        <div>
          <h1 class="text-h4 font-weight-bold mb-2">Notifications</h1>
          <p class="text-body-1 text-medium-emphasis">
            Manage and view all your notifications
          </p>
        </div>

        <div class="d-flex gap-3">
          <v-btn
            v-if="unreadCount > 0"
            variant="outlined"
            color="primary"
            @click="markAllAsRead"
            :loading="loading"
            prepend-icon="mdi-check-all"
          >
            Mark All Read
          </v-btn>

          <v-btn
            variant="outlined"
            color="secondary"
            @click="confirmClearAll"
            :loading="loading"
            prepend-icon="mdi-delete-sweep"
          >
            Clear All
          </v-btn>
        </div>
      </div>

      <!-- Filters -->
      <v-card class="mb-6" elevation="2">
        <v-card-text class="pa-4">
          <div class="d-flex align-center gap-4">
            <v-btn-toggle
              v-model="statusFilter"
              mandatory
              color="primary"
              variant="outlined"
            >
              <v-btn value="all">All</v-btn>
              <v-btn value="unread">Unread ({{ unreadCount }})</v-btn>
              <v-btn value="read">Read</v-btn>
            </v-btn-toggle>

            <v-select
              v-model="typeFilter"
              :items="typeOptions"
              label="Type"
              variant="outlined"
              density="compact"
              hide-details
              class="ml-4"
              style="min-width: 200px"
            ></v-select>
          </div>
        </v-card-text>
      </v-card>

      <!-- Notifications List -->
      <v-card elevation="2">
        <v-card-text class="pa-0">
          <div v-if="loading && persistentNotifications.length === 0" class="pa-8 text-center">
            <v-progress-circular indeterminate color="primary" size="64"></v-progress-circular>
            <div class="mt-4 text-h6 text-medium-emphasis">Loading notifications...</div>
          </div>

          <div v-else-if="filteredNotifications.length === 0" class="pa-8 text-center">
            <v-icon icon="mdi-bell-off" size="80" color="grey-lighten-1"></v-icon>
            <div class="mt-4 text-h6 text-medium-emphasis">No notifications found</div>
            <p class="text-body-1 text-grey mt-2">
              {{ statusFilter === 'all' ? 'You have no notifications yet.' : `No ${statusFilter} notifications found.` }}
            </p>
          </div>

          <div v-else>
            <v-list class="pa-0">
              <v-list-item
                v-for="notification in paginatedNotifications"
                :key="notification.id"
                :class="{ 
                  'bg-blue-lighten-5 border-s-lg border-primary': notification.status === 'unread',
                  'notification-clickable': true
                }"
                @click="onNotificationClick(notification)"
                class="pa-4 mb-2 mx-2 rounded-lg elevation-1"
                :ripple="true"
              >
                <template v-slot:prepend>
                  <v-avatar
                    size="48"
                    :color="getNotificationColor(notification.type)"
                    class="notification-icon mr-4"
                  >
                    <v-icon
                      :icon="getNotificationIcon(notification.type)"
                      color="white"
                      size="24"
                    ></v-icon>
                  </v-avatar>
                </template>

                <v-list-item-title class="text-h6 font-weight-medium mb-2 text-primary notification-title">
                  {{ notification.title }}
                </v-list-item-title>

                <v-list-item-subtitle class="text-body-1 mb-2 text-medium-emphasis notification-message">
                  {{ notification.message }}
                </v-list-item-subtitle>

                <div class="d-flex align-center justify-space-between">
                  <v-list-item-subtitle class="notification-time text-caption text-medium-emphasis">
                    {{ formatNotificationTime(notification.created_at) }}
                  </v-list-item-subtitle>

                  <div class="d-flex gap-2">
                    <v-chip
                      :color="getNotificationColor(notification.type)"
                      variant="tonal"
                      size="small"
                      class="text-caption"
                    >
                      {{ notification.type.replace('_', ' ').toUpperCase() }}
                    </v-chip>

                    <v-chip
                      v-if="notification.status === 'unread'"
                      color="primary"
                      variant="tonal"
                      size="small"
                      class="text-caption"
                    >
                      UNREAD
                    </v-chip>
                  </div>
                </div>
              </v-list-item>
            </v-list>

            <!-- Pagination -->
            <v-divider></v-divider>
            <div class="d-flex align-center justify-space-between pa-4">
              <div class="text-body-2 text-medium-emphasis">
                Showing {{ startIndex + 1 }}-{{ endIndex }} of {{ filteredNotifications.length }} notifications
              </div>

              <v-pagination
                v-model="currentPage"
                :length="totalPages"
                :total-visible="7"
                color="primary"
              ></v-pagination>
            </div>
          </div>
        </v-card-text>
      </v-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useNotificationStore, type Notification } from '../stores/notifications'
import { getNotificationIcon, getNotificationColor, formatNotificationTime } from '../utils/notificationUtils'

const notificationStore = useNotificationStore()
const {
  persistentNotifications,
  loading,
  unreadCount,
  markAllAsRead,
  clearAll,
  loadNotifications
} = notificationStore

// Filters
const statusFilter = ref<'all' | 'read' | 'unread'>('all')
const typeFilter = ref<'all' | 'deployment_update' | 'task_update' | 'connected' | 'error'>('all')
const currentPage = ref(1)
const pageSize = 10

// Type options for filter
const typeOptions = computed(() => [
  { title: 'All Types', value: 'all' },
  { title: 'Deployment Updates', value: 'deployment_update' },
  { title: 'Task Updates', value: 'task_update' },
  { title: 'Connection Events', value: 'connected' },
  { title: 'Errors', value: 'error' }
])

// Filtered notifications
const filteredNotifications = computed(() => {
  let filtered = [...persistentNotifications]

  // Filter by status
  if (statusFilter.value !== 'all') {
    filtered = filtered.filter((n: Notification) => n.status === statusFilter.value as 'read' | 'unread')
  }

  // Filter by type
  if (typeFilter.value !== 'all') {
    filtered = filtered.filter((n: Notification) => n.type === typeFilter.value)
  }

  return filtered
})

// Pagination
const totalPages = computed(() => Math.ceil(filteredNotifications.value.length / pageSize))
const startIndex = computed(() => (currentPage.value - 1) * pageSize)
const endIndex = computed(() => Math.min(startIndex.value + pageSize, filteredNotifications.value.length))
const paginatedNotifications = computed(() =>
  filteredNotifications.value.slice(startIndex.value, endIndex.value)
)

// Methods
const onNotificationClick = async (notification: Notification) => {
  if (notification.status === 'unread') {
    await notificationStore.markAsRead(notification.id)
  }
}

const confirmClearAll = () => {
  if (confirm('Are you sure you want to clear all notifications? This action cannot be undone.')) {
    clearAll()
  }
}

// Watch for filter changes
watch([statusFilter, typeFilter], () => {
  currentPage.value = 1
})

// Initial load
onMounted(() => {
  if (persistentNotifications.length === 0) {
    loadNotifications()
  }
})
</script>

<style scoped>
.notifications-page {
  min-height: 100vh;
  background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%), radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%);
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .container {
    padding: 16px;
  }

  .d-flex.align-center.justify-space-between {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .d-flex.gap-3 {
    flex-direction: column;
    gap: 8px;
  }
}

.notification-clickable {
  cursor: pointer;
  transition: all 0.2s ease;
}

.notification-title {
  word-wrap: break-word;
  overflow-wrap: break-word;
  white-space: normal;
  line-height: 1.4;
}

.notification-message {
  word-wrap: break-word;
  overflow-wrap: break-word;
  white-space: normal;
  line-height: 1.5;
  max-width: none;
}
</style>