<template>
  <div class="requests-table-container">
    <v-data-table :loading="loading" :headers="headers" :items="pendingRequests" class="requests-table"
      :items-per-page="5" :no-data-text="'No pending requests found'" density="comfortable">
      <template v-if="showUserID" v-slot:[`item.user_id`]="{ item }">
        <slot name="user_id" :item="item"></slot>
      </template>
      <template v-slot:[`item.created_at`]="{ item }">
        <span>{{ formatDate(item.created_at) }}</span>
      </template>
      <template v-slot:[`item.usd_amount`]="{ item }">
        <span>${{ item.usd_amount.toFixed(2) }}</span>
      </template>
      <template v-slot:[`item.transferred_usd_amount`]="{ item }">
        <span>${{ item.transferred_usd_amount.toFixed(2) }}</span>
      </template>
      <template v-slot:[`item.status`]="{ item }">
        <v-chip :color="getStatusColor(item)" size="small" class="status-chip">
          {{ getStatus(item) }}
        </v-chip>
      </template>

    </v-data-table>
  </div>
</template>

<script setup lang="ts">
import { type PendingRequest } from '../../utils/userService'
import { formatDate } from '../../utils/uiUtils.ts'

const props = defineProps({
  pendingRequests: {
    type: Array as () => PendingRequest[],
    required: true,
    default: () => []
  },
  showUserID: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const headers = [
  { title: 'Request Date', key: 'created_at' },
  { title: 'Requested Amount', key: 'usd_amount' },
  { title: 'Transferred Amount', key: 'transferred_usd_amount' },
  { title: 'Status', key: 'status' },
]

if (props.showUserID) {
  headers.unshift({ title: 'User ID', key: 'user_id' })
}



function getStatus(item: PendingRequest): string {
  if (item.transferred_usd_amount === item.usd_amount &&
    item.transferred_tft_amount === item.tft_amount) {
    return 'Completed'
  }
  return 'Pending'
}

function getStatusColor(item: PendingRequest): string {
  const status = getStatus(item)
  if (status === 'Completed') {
    return 'success'
  }
  return 'gray'
}
</script>

<style scoped>
.requests-table-container {
  margin-bottom: var(--space-6);
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.requests-table {
  background: transparent;
  width: 100%;
}

.status-chip {
  font-size: 0.75rem;
  font-weight: var(--font-weight-medium);
}
</style>
