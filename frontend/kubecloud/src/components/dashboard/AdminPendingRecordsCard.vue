<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Pending Records</h3>
          <p class="dashboard-card-subtitle">View users pending transfer records</p>
        </div>
      </div>
    </div>
    <PendingRecordsTable 
      :pendingRecords="pendingRecords" 
      :showUserID="true"
      :loading="loading"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { type PendingRecord } from '../../utils/userService'
import PendingRecordsTable from './PendingRecordsTable.vue'
import { useNotificationStore } from '@/stores/notifications'
import { adminService } from '@/utils/adminService'

const pendingRecords = ref<PendingRecord[]>([])
const notificationStore = useNotificationStore()

onMounted(async () => {
  await loadPendingRecords()
})

const loading = ref(false)

async function loadPendingRecords() {
  loading.value = true
  try {
    const response = await adminService.listPendingRecords()
    pendingRecords.value = response || []
  } catch (error) {
    console.error('Failed to load pending records:', error)
    notificationStore.error('Error', 'Failed to load pending records')
  } finally {
    loading.value = false
  }
}


</script>

<style scoped>
/* Card styling is inherited from global styles */
</style>
