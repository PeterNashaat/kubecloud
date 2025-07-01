<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Add Funds</h3>
          <p class="dashboard-card-subtitle">Add funds to your account balance</p>
        </div>
      </div>
    </div>
    <div class="dashboard-card-content">
      <div class="balance-section list-item-interactive">
        <span class="balance-label">Current Balance:</span>
        <span class="balance-value">$ {{ user?.credited_balance ?? 0 }}</span>
      </div>
      <div class="amount-section">
        <label class="section-label">Amount</label>
        <div class="amount-options">
          <button
            v-for="preset in presets"
            :key="preset"
            :class="['amount-btn', { selected: amount === preset }]"
            @click="selectAmount(preset)"
          >
            {{ preset }}
          </button>
          <div class="custom-amount-wrapper">
            <input
              type="number"
              min="1"
              class="amount-input"
              v-model.number="customAmount"
              @focus="selectAmount('custom')"
              :class="{ selected: amount === 'custom' }"
              placeholder="Custom"
            />
          </div>
        </div>
      </div>
      <div class="card-details-section">
        <label class="section-label">Card Details</label>
        <div class="card-details-fields">
          <input
            class="card-input"
            type="text"
            maxlength="19"
            placeholder="Card Number"
            v-model="cardNumber"
          />
          <input
            class="card-input short"
            type="text"
            maxlength="5"
            placeholder="mm/dd"
            v-model="expiry"
          />
          <input
            class="card-input short"
            type="text"
            maxlength="4"
            placeholder="CVC"
            v-model="cvc"
          />
        </div>
      </div>
      <div class="charge-section">
        <button class="action-btn charge-btn" @click="chargeBalance" :disabled="loading">
          Charge Balance
        </button>
      </div>
      <div v-if="successMessage" class="success-message">{{ successMessage }}</div>
      <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import { userService } from '../../utils/userService'

const { user } = storeToRefs(useUserStore())

const presets = [5, 10, 20, 50]
const amount = ref<number | 'custom'>(5)
const customAmount = ref<number | null>(null)
const cardNumber = ref('')
const expiry = ref('')
const cvc = ref('')
const loading = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

function selectAmount(val: number | 'custom') {
  amount.value = val
  if (val !== 'custom') customAmount.value = null
}

async function chargeBalance() {
  loading.value = true
  successMessage.value = ''
  errorMessage.value = ''
  const selectedAmount = getSelectedAmount()
  if (!selectedAmount || !cardNumber.value || !expiry.value || !cvc.value) {
    errorMessage.value = 'Please fill all fields.'
    loading.value = false
    return
  }
  try {
    // Compose payment_method_id as a string (could be card number + expiry + cvc for demo)
    const payment_method_id = `${cardNumber.value}|${expiry.value}|${cvc.value}`
    const payload = {
      card_type: 'credit', // or detect from card number
      payment_method_id,
      amount: Number(selectedAmount)
    }
    await userService.chargeBalance(payload)
    successMessage.value = 'Balance charged successfully!'
    // Optionally update balance here
  } catch (err: any) {
    errorMessage.value = err?.message || 'Failed to charge balance.'
  } finally {
    loading.value = false
  }
}

function getSelectedAmount() {
  return amount.value === 'custom' ? customAmount.value : amount.value
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
.balance-section {
  display: flex;
  gap: 0.5rem;
  margin-bottom: var(--space-2);
  padding: var(--space-4);
  border-radius: var(--radius-lg);
  background: rgba(30, 41, 59, 0.7);
  border: 1px solid var(--color-border);
  transition: background 0.18s, border-color 0.18s;
}
.balance-section.list-item-interactive:hover {
  background: rgba(30, 41, 59, 0.85);
  border-color: var(--color-border-light);
}
.balance-label {
  font-weight: 500;
  color: var(--color-text, #CBD5E1);
}
.balance-value {
  font-weight: 700;
  color: var(--color-success, #10B981);
  font-size: 1.2rem;
}
.amount-section {
  margin-bottom: 0;
}
.section-label {
  font-weight: 500;
  color: var(--color-text, #CBD5E1);
  margin-bottom: 0.5rem;
  display: block;
}
.amount-options {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
}
.amount-btn {
  background: var(--color-bg-btn, #232f47);
  border: 1.5px solid var(--color-border, #334155);
  border-radius: 0.75rem;
  padding: 0.5rem 1.2rem;
  font-size: 1rem;
  color: var(--color-text);
  cursor: pointer;
  font-weight: 500;
  transition: background 0.15s, border-color 0.15s, color 0.15s;
  outline: none;
}
.amount-input {
  width: 7rem;
  padding: 0.5rem 0.7rem;
  border: 1.5px solid var(--color-border, #334155);
  border-radius: 0.75rem;
  font-size: 1rem;
  color: var(--color-text);
  background: var(--color-bg-btn, #232f47);
  outline: none;
  font-weight: 500;
  margin-left: 0.2rem;
}
.card-details-section {
  margin-bottom: 0;
}
.card-details-fields {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
}
.card-input {
  padding: 0.5rem 1rem;
  border: 1.5px solid var(--color-border, #334155);
  border-radius: 0.75rem;
  font-size: 1rem;
  color: var(--color-text);
  background: var(--color-bg-btn, #232f47);
  outline: none;
  font-weight: 500;
}
.card-input.short {
  width: 5.5rem;
}
.charge-section {
  margin-top: 0.5rem;
}
.charge-btn {
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-text);
  font-weight: var(--font-weight-medium);
  border-radius: 0.75rem;
  padding: 0.7rem 2.5rem;
  font-size: 1.1rem;
  cursor: pointer;
  transition: background 0.18s, border-color 0.18s;
  box-shadow: none;
}
.charge-btn:hover {
  background: rgba(59, 130, 246, 0.07);
  border-color: var(--color-primary);
  color: var(--color-primary);
}
.charge-btn:disabled {
  background: #38BDF899;
  cursor: not-allowed;
}
.success-message {
  color: var(--color-success, #10B981);
  margin-top: 1.2rem;
  font-weight: 500;
}
.error-message {
  color: var(--color-error, #ef4444);
  margin-top: 1.2rem;
  font-weight: 500;
}
</style>
