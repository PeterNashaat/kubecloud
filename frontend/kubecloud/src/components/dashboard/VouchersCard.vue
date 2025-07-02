<template>
  <div class="dashboard-card vouchers-card spacious">
    <div class="dashboard-card-header">
      <h3 class="dashboard-card-title">Redeem Voucher</h3>
      <p class="dashboard-card-subtitle">Add credits to your balance using a voucher code</p>
    </div>
    <div class="dashboard-card-content">
      <input
        v-model="code"
        class="voucher-input bordered"
        type="text"
        placeholder="Voucher Code"
        :disabled="loading"
        @keyup.enter="onRedeem"
      />
      <v-btn
        color="primary"
        :loading="loading"
        :disabled="loading"
        class="redeem-btn action-btn"
        @click="onRedeem"
        prepend-icon="mdi-gift"
      >
        Redeem
      </v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

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

function onRedeem() {
  if (!code.value.trim()) {
    // For demo, treat any code as success
    code.value = ''
  }
  loading.value = true
  // Simulate async redeem
  setTimeout(() => {
    loading.value = false
    // For demo, treat any code as success
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
.dashboard-card.vouchers-card.spacious {
  background: #181f35;
  border-radius: 1.25rem;
  border: 1.5px solid #334155;
  max-width: 520px;
  margin: 0;
  padding: 2.2rem 2.5rem 2.2rem 2.5rem;
  box-shadow: 0 4px 32px 0 rgba(0,0,0,0.12);
  display: flex;
  flex-direction: column;
}
.dashboard-card-header {
  margin-bottom: 1.5rem;
  width: 100%;
}
.dashboard-card-title {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.2rem;
}
.dashboard-card-subtitle {
  font-size: 1.05rem;
  color: #60a5fa;
  margin-bottom: 0.5rem;
}
.dashboard-card-content {
  width: 100%;
}
.voucher-input {
  width: 100%;
  padding: 0.7rem 1.1rem;
  font-size: 1.1rem;
  background: #232f47;
  color: #CBD5E1;
  margin-bottom: 1.3rem;
}
.voucher-input.bordered {
  border: 1.5px solid #334155;
  border-radius: 0.7rem;
  outline: none;
  transition: border-color 0.15s;
}
.voucher-input.bordered:focus {
  border-color: #60a5fa;
}
.redeem-btn.action-btn {
  margin-top: 1.2rem;
  font-size: 1.1rem;
  padding: 0.9rem 0;
  border-radius: 0.7rem;
}
</style>
