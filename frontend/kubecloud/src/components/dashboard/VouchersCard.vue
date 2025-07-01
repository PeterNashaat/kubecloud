<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Redeem Voucher</h3>
          <p class="dashboard-card-subtitle">Add credits to your balance using a voucher code</p>
        </div>
      </div>
    </div>
    <div class="dashboard-card-content">
      <label class="voucher-label" for="voucher-code">Code</label>
      <input
        id="voucher-code"
        v-model="code"
        class="voucher-input"
        type="text"
        placeholder="Enter voucher code"
        @keyup.enter="onRedeem"
      />
      <button class="action-btn redeem-btn" @click="onRedeem">Redeem</button>
      <div v-if="successMessage" class="success-message">{{ successMessage }}</div>
      <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const code = ref('')
const successMessage = ref('')
const errorMessage = ref('')

function onRedeem() {
  successMessage.value = ''
  errorMessage.value = ''
  if (!code.value.trim()) {
    errorMessage.value = 'Please enter a code.'
    return
  }
  emit('redeem', code.value)
}

const emit = defineEmits(['redeem'])
</script>

<script lang="ts">
export default {
  name: 'VouchersCard'
}
</script>

<style scoped>
.dashboard-card {
  background: var(--color-bg-card, #182235);
  border-radius: var(--radius-xl, 1.25rem);
  box-shadow: 0 2px 16px 0 rgba(0,0,0,0.08);
  padding: var(--space-8);
  margin-bottom: var(--space-8);
  border: 1px solid var(--color-border, #334155);
  max-width: 480px;
  margin: 0;
  padding: 0 !important;
}
.dashboard-card-header {
  margin-bottom: var(--space-6);
}
.dashboard-card-title {
  font-size: var(--font-size-xl, 1.5rem);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text, #fff);
  margin: 0 0 var(--space-2) 0;
}
.dashboard-card-subtitle {
  font-size: var(--font-size-base, 1rem);
  color: var(--color-primary, #38BDF8);
  font-weight: var(--font-weight-medium, 500);
  opacity: 0.9;
  margin: 0;
}
.dashboard-card-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  align-items: flex-start;
}
.voucher-label {
  font-size: 1rem;
  font-weight: 500;
  margin-bottom: 0.5rem;
  color: var(--color-text, #fff);
}
.voucher-input {
  width: 320px;
  padding: 0.75rem 1rem;
  font-size: 1.1rem;
  border-radius: 0.75rem;
  border: 1px solid var(--color-border, #334155);
  margin-bottom: 1.5rem;
  outline: none;
}
.action-btn.redeem-btn {
  width: 180px;
  padding: 0.7rem 0;
  font-size: 1.1rem;
  font-weight: 600;
  border-radius: 0.75rem;
  border: 1px solid var(--color-border, #334155);
  background: transparent;
  color: var(--color-text, #fff);
  cursor: pointer;
  margin-bottom: 1.5rem;
  transition: background 0.18s, border-color 0.18s;
}
.action-btn.redeem-btn:hover {
  background: rgba(59, 130, 246, 0.07);
  border-color: var(--color-primary, #3B82F6);
  color: var(--color-primary, #3B82F6);
}
.success-message {
  color: var(--color-success, #10B981);
  font-weight: 500;
  margin-top: 0.5rem;
}
.error-message {
  color: var(--color-error, #ef4444);
  font-weight: 500;
  margin-top: 0.5rem;
}
</style>
