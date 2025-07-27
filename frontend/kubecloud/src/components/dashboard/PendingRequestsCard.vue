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
    <div class="requests-table-container">
      <v-data-table
        :headers="headers"
        :items="pendingRequests"
        class="requests-table"
        :items-per-page="5"
        :no-data-text="'No pending requests found'"
        density="comfortable"
        :hide-default-footer="false"
      >
        <template v-slot:[`item.created_at`]="{ item }">
          <span>{{ formatDate(item.created_at) }}</span>
        </template>
        <template v-slot:[`item.usd_amount`]="{ item }">
          <span>${{ item.usd_amount.toFixed(2) }}</span>
        </template>
        <template v-slot:[`item.tft_amount`]="{ item }">
          <span>{{ item.tft_amount.toFixed(2) }} TFT</span>
        </template>
        <template v-slot:[`item.transferred_usd_amount`]="{ item }">
          <span>${{ item.transferred_usd_amount.toFixed(2) }}</span>
        </template>
        <template v-slot:[`item.transferred_tft_amount`]="{ item }">
          <span>{{ item.transferred_tft_amount.toFixed(2) }} TFT</span>
        </template>
        <template v-slot:[`item.status`]="{ item }">
          <v-chip
            :color="getStatusColor(item)"
            size="small"
            class="status-chip"
          >
            {{ getStatus(item) }}
          </v-chip>
        </template>
      </v-data-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'
import type { PendingRequest } from '../../utils/userService'

const props = defineProps<{ pendingRequests: PendingRequest[] }>()

const headers = [
  { title: 'Date', key: 'created_at' },
  { title: 'USD Amount', key: 'usd_amount' },
  { title: 'TFT Amount', key: 'tft_amount' },
  { title: 'Transferred USD', key: 'transferred_usd_amount' },
  { title: 'Transferred TFT', key: 'transferred_tft_amount' },
  { title: 'Status', key: 'status', sortable: false },
]

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

function getStatus(item: PendingRequest): string {
  // If transferred amounts match requested amounts, it's completed
  if (item.transferred_usd_amount === item.usd_amount && 
      item.transferred_tft_amount === item.tft_amount) {
    return 'Completed'
  }
  
  // If any amount is transferred but not all, it's partially completed
  if (item.transferred_usd_amount > 0 || item.transferred_tft_amount > 0) {
    return 'Partial'
  }
  
  // Otherwise it's pending
  return 'Pending'
}

function getStatusColor(item: PendingRequest): string {
  const status = getStatus(item)
  
  switch (status) {
    case 'Completed':
      return 'success'
    case 'Partial':
      return 'warning'
    case 'Pending':
      return 'info'
    default:
      return 'grey'
  }
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
