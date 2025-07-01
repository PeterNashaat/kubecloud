<template>
  <div class="dashboard-card vouchers-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Redeem Voucher</h3>
          <p class="dashboard-card-subtitle">Add credits to your balance using a voucher code</p>
        </div>
      </div>
    </div>
    <div class="dashboard-card-content">
      <v-text-field
        v-model="code"
        label="Voucher Code"
        prepend-inner-icon="mdi-ticket-percent"
        variant="outlined"
        :disabled="loading"
        @keyup.enter="onRedeem"
        class="voucher-input-field"
      />
      <v-btn
        color="primary"
        :loading="loading"
        :disabled="loading"
        class="redeem-btn"
        @click="onRedeem"
        prepend-icon="mdi-gift"
      >
        Redeem
      </v-btn>
      <v-alert v-if="successMessage" type="success" variant="tonal" class="mt-3" border="start" icon="mdi-check-circle">
        {{ successMessage }}
      </v-alert>
      <v-alert v-if="errorMessage" type="error" variant="tonal" class="mt-3" border="start" icon="mdi-alert-circle">
        {{ errorMessage }}
      </v-alert>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { VTextField, VBtn, VAlert } from 'vuetify/components'

interface Voucher {
  id: number | string
  name: string
  description?: string
  amount: string
  expiryDate: string
  used?: boolean
  icon?: string
  iconColor?: string
}

const props = defineProps<{ vouchers: Voucher[] }>()

const code = ref('')
const loading = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

function onRedeem() {
  successMessage.value = ''
  errorMessage.value = ''
  if (!code.value.trim()) {
    errorMessage.value = 'Please enter a code.'
    return
  }
  loading.value = true
  // Simulate async redeem
  setTimeout(() => {
    loading.value = false
    // For demo, treat any code as success
    successMessage.value = 'Voucher redeemed successfully!'
    code.value = ''
  }, 1200)
  emit('redeem', code.value)
}

function isExpired(expiryDate: string) {
  if (!expiryDate) return false
  const now = new Date()
  const exp = new Date(expiryDate)
  return exp < now
}

const emit = defineEmits(['redeem'])
</script>

<script lang="ts">
export default {
  name: 'VouchersCard'
}
</script>

<style scoped>
/* Remove centering and use full width like other dashboard cards */
.vouchers-card {
  width: 100%;
  max-width: unset;
  margin: 0;
}
.voucher-input-field {
  width: 100%;
  max-width: 340px;
}
.redeem-btn {
  min-width: 160px;
  margin-top: 0.5rem;
}
</style>
