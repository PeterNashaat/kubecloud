<template>
  <div class="admin-section">
    <div class="section-header">
      <h2 class="dashboard-title">Voucher Management</h2>
      <p class="section-subtitle">Generate and manage platform vouchers</p>
    </div>
    
    <div class="dashboard-card">
      <div class="dashboard-card-header">
        <h3 class="dashboard-card-title">Generate Vouchers</h3>
        <p class="dashboard-card-subtitle">Create new vouchers for user promotions</p>
      </div>
      
      <v-form @submit.prevent="handleGenerateVouchers" class="voucher-form">
        <div class="form-row">
          <v-text-field
            v-model.number="form.voucherValue"
            label="Voucher Value ($)"
            type="number"
            prepend-inner-icon="mdi-currency-usd"
            variant="outlined"
            min="1"
            max="10000"
            density="comfortable"
            required
            :rules="[rules.required, rules.range(1, 10000)]"
            class="form-field"
            :disabled="isGenerating"
          />
          <v-text-field
            v-model.number="form.voucherCount"
            label="Number of Vouchers"
            type="number"
            prepend-inner-icon="mdi-pound"
            variant="outlined"
            min="1"
            max="1000"
            density="comfortable"
            required
            :rules="[rules.required, rules.range(1, 1000)]"
            class="form-field"
            :disabled="isGenerating"
          />
          <v-text-field
            v-model.number="form.voucherExpiry"
            label="Expiry (days)"
            type="number"
            prepend-inner-icon="mdi-calendar-clock"
            variant="outlined"
            min="1"
            max="365"
            density="comfortable"
            required
            :rules="[rules.required, rules.range(1, 365)]"
            class="form-field"
            :disabled="isGenerating"
          />
        </div>
        
        <v-btn 
          type="submit" 
          color="primary" 
          variant="elevated" 
          class="btn-primary"
          :loading="isGenerating"
          :disabled="!isFormValid"
        >
          <v-icon icon="mdi-ticket-percent" class="mr-2"></v-icon>
          {{ isGenerating ? 'Generating...' : 'Generate' }}
        </v-btn>
      </v-form>
    </div>
    
    <div class="dashboard-card">
      <div class="dashboard-card-header">
        <h3 class="dashboard-card-title">All Vouchers</h3>
        <p class="dashboard-card-subtitle">List of all generated vouchers</p>
      </div>
      
      <div class="table-container">
        <v-data-table
          :headers="tableHeaders"
          :items="paginatedVouchers"
          class="admin-table"
          hide-default-footer
          density="comfortable"
        >
          <template #item.redeemed="{ item }">
            {{ item.redeemed ? 'Yes' : 'No' }}
          </template>
          <template #item.created_at="{ item }">
            {{ formatDate(item.created_at) }}
          </template>
          <template #item.expires_at="{ item }">
            {{ formatDate(item.expires_at) }}
          </template>
        </v-data-table>
        
        <div class="pagination-container mt-4">
          <v-pagination
            v-model="currentPage"
            :length="totalPages"
            color="primary"
            circle
            size="small"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps({
  vouchers: { type: Array as () => any[], default: () => [] }
})

const emit = defineEmits(['generateVouchers'])

// Consolidated form state
const form = ref({ voucherValue: 50, voucherCount: 10, voucherExpiry: 30 })
const isGenerating = ref(false)
const currentPage = ref(1)

// Minimal validation rules
const rules = {
  required: (value: any) => !!value || 'This field is required',
  range: (min: number, max: number) => (value: number) => 
    (value >= min && value <= max) || `Value must be between ${min} and ${max}`
}

// Simple form validation
const isFormValid = computed(() => 
  Object.values(form.value).every(val => val >= 1) &&
  form.value.voucherValue <= 10000 &&
  form.value.voucherCount <= 1000 &&
  form.value.voucherExpiry <= 365
)

// Table configuration
const tableHeaders = [
  { title: 'Voucher', key: 'code' },
  { title: 'Value', key: 'value' },
  { title: 'Redeemed', key: 'redeemed' },
  { title: 'Created At', key: 'created_at' },
  { title: 'Expires At', key: 'expires_at' }
]

// Pagination
const pageSize = 10
const totalPages = computed(() => Math.ceil(props.vouchers.length / pageSize))
const paginatedVouchers = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return props.vouchers.slice(start, start + pageSize)
})

// Form submission
const handleGenerateVouchers = async () => {
  if (!isFormValid.value) return
  
  try {
    isGenerating.value = true
    await emit('generateVouchers', {
      count: form.value.voucherCount,
      value: form.value.voucherValue,
      expire_after_days: form.value.voucherExpiry
    })
    
    // Reset form
    Object.assign(form.value, { voucherValue: 50, voucherCount: 10, voucherExpiry: 30 })
  } catch (error) {
    console.error('Failed to generate vouchers:', error)
  } finally {
    isGenerating.value = false
  }
}

// Utility function
const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return isNaN(d.getTime()) ? dateStr : 
    d.toLocaleString(undefined, { 
      year: 'numeric', month: '2-digit', day: '2-digit', 
      hour: '2-digit', minute: '2-digit' 
    })
}
</script>
