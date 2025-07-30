<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Pending Requests</h3>
          <p class="dashboard-card-subtitle">View users pending transfer requests</p>
        </div>
      </div>
    </div>
    <PendingRequestsTable 
      :pendingRequests="pendingRequests" 
      :showUserID="true"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { type PendingRequest } from '../../utils/userService'
import PendingRequestsTable from './PendingRequestsTable.vue'
import { useNotificationStore } from '@/stores/notifications'
import { adminService } from '@/utils/adminService'

const pendingRequests = ref<PendingRequest[]>([])
const notificationStore = useNotificationStore()

onMounted(async () => {
  await loadPendingRequests()
})

async function loadPendingRequests() {
  try {
    const response = await adminService.listPendingRequests()
    pendingRequests.value = response || []
  } catch (error) {
    console.error('Failed to load pending requests:', error)
    notificationStore.error('Error', 'Failed to load pending requests')
  }
}


</script>

<style scoped>
/* Card styling is inherited from global styles */
</style>
