<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Pending Requests</h3>
          <p class="dashboard-card-subtitle">View your pending transfer requests</p>
        </div>
      </div>
    </div>
    <PendingRequestsTable :pendingRequests="pendingRequests" :loading="loading" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { userService, type PendingRequest } from '../../utils/userService'
import PendingRequestsTable from './PendingRequestsTable.vue'

const pendingRequests = ref<PendingRequest[]>([])
const loading = ref(true)
onMounted(async () => {
  try {
    loading.value = true
    const response = await userService.listUserPendingRequests()
    pendingRequests.value = response || []
    loading.value = false
  } catch (error) {
    console.error('Failed to load user pending requests:', error)
    loading.value = false
  }
})
</script>

<style scoped>
/* Card styling is inherited from global styles */
</style>
