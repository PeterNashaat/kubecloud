<template>
  <div class="dashboard-card payment-card spacious">
    <div class="dashboard-card-header">
      <h3 class="dashboard-card-title">Add Funds</h3>
      <p class="dashboard-card-subtitle">Add funds to your account balance</p>
    </div>
    <div class="dashboard-card-content">
      <div class="balance-row">
        <span>Current Balance:</span>
        <span class="balance-value">${{ user?.credited_balance ?? 0 }}</span>
      </div>
      <div class="amount-row">
        <span>Amount:</span>
        <div class="amount-options">
          <button
            v-for="preset in presets"
            :key="preset"
            :class="{ selected: amount === preset }"
            @click="selectAmount(preset)"
            type="button"
          >{{ preset }}</button>
          <template v-if="amount !== 'custom'">
            <button
              :class="{ selected: amount === 'custom' }"
              @click="selectAmount('custom')"
              type="button"
            >Custom</button>
          </template>
          <input
            v-if="amount === 'custom'"
            type="number"
            min="1"
            class="amount-input"
            v-model.number="customAmount"
            placeholder="Custom"
            @focus="selectAmount('custom')"
          />
        </div>
      </div>
      <div class="card-details-row">
        <input
          type="text"
          class="card-input bordered"
          v-model="cardNumber"
          placeholder="Card Number"
          maxlength="19"
        />
        <input
          type="text"
          class="card-input short bordered"
          v-model="expiryDate"
          placeholder="MM/YY"
          maxlength="5"
          @input="formatExpiryDate"
        />
        <input
          type="text"
          class="card-input short bordered"
          v-model="cvv"
          placeholder="CVV"
          maxlength="4"
        />
      </div>
      <v-btn
        class="action-btn"
        color="primary"
        :loading="loading"
        :disabled="loading || !isFormValid"
        @click="chargeBalance"
        prepend-icon="mdi-credit-card-plus"
      >
        Charge Balance
      </v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useUserStore } from '../../stores/user'
import { userService } from '../../utils/userService'
import { api } from '../../utils/api'
import type { ApiResponse } from '../../utils/authService'
import type { User } from '../../stores/user'

const { user } = storeToRefs(useUserStore())

const presets = [5, 10, 20, 50]
const amount = ref<number | 'custom'>(5)
const customAmount = ref<number | null>(null)
const loading = ref(false)

// Card details
const cardNumber = ref('')
const expiryDate = ref('')
const cvv = ref('')

// Form validation
const isFormValid = computed(() => {
  const selectedAmount = getSelectedAmount()
  return selectedAmount && selectedAmount > 0 && 
         cardNumber.value.replace(/\s/g, '').length >= 13 &&
         expiryDate.value.length === 5 &&
         cvv.value.length >= 3
})

function selectAmount(val: number | 'custom') {
  amount.value = val
  if (val !== 'custom') customAmount.value = null
}

function formatCardNumber() {
  let value = cardNumber.value.replace(/\s/g, '')
  value = value.replace(/\D/g, '')
  value = value.replace(/(\d{4})(?=\d)/g, '$1 ')
  cardNumber.value = value
}

function formatExpiryDate() {
  let value = expiryDate.value.replace(/\D/g, '')
  if (value.length >= 2) {
    value = value.substring(0, 2) + '/' + value.substring(2, 4)
  }
  expiryDate.value = value
}

async function chargeBalance() {
  loading.value = true
  
  const selectedAmount = getSelectedAmount()
  if (!selectedAmount || !isFormValid.value) {
    return
  }

  try {
    // For demo purposes, we'll use a mock payment method ID
    // In production, this would be created through Stripe
    const mockPaymentMethodId = 'pm_' + Math.random().toString(36).substr(2, 9)
    
    await userService.chargeBalance({
      card_type: 'card',
      payment_method_id: mockPaymentMethodId,
      amount: Number(selectedAmount)
    })
    
    // Clear the form
    cardNumber.value = ''
    expiryDate.value = ''
    cvv.value = ''
    amount.value = 5
    customAmount.value = null
    
    // Refresh user data to get updated balance
    const userStore = useUserStore()
    try {
      const userRes = await api.get<ApiResponse<{ user: User }>>('/v1/user/', { requiresAuth: true, showNotifications: false })
      userStore.user = userRes.data.data.user
    } catch (error) {
      console.error('Failed to refresh user data:', error)
    }
    
  } catch (err: any) {
    console.error('Failed to charge balance:', err)
  } finally {
    loading.value = false
  }
}

function getSelectedAmount() {
  return amount.value === 'custom' ? customAmount.value : amount.value
}
</script>

<style scoped>
.dashboard-card.payment-card.spacious {
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
.balance-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.3rem;
  font-size: 1.1rem;
}
.balance-value {
  color: #10B981;
  font-weight: 700;
  font-size: 1.2rem;
}
.amount-row {
  display: flex;
  align-items: center;
  margin-bottom: 1.3rem;
  gap: 0.7rem;
}
.amount-options {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.amount-btn,
.amount-options button {
  background: #232f47;
  border: 1.5px solid #334155;
  border-radius: 0.7rem;
  padding: 0.5rem 1.3rem;
  font-size: 1.1rem;
  color: #CBD5E1;
  cursor: pointer;
  font-weight: 600;
  transition: background 0.15s, border-color 0.15s, color 0.15s;
}
.amount-options button.selected,
.amount-options button:focus {
  background: #60a5fa;
  color: #fff;
  border-color: #60a5fa;
}
.amount-input {
  width: 100px;
  padding: 0.5rem 0.7rem;
  border: 1.5px solid #334155;
  border-radius: 0.7rem;
  font-size: 1.1rem;
  margin-left: 0.3rem;
}
.card-details-row {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.3rem;
}
.card-input {
  padding: 0.7rem 1.1rem;
  font-size: 1.1rem;
  min-width: 0;
  width: 200px;
  background: #232f47;
  color: #CBD5E1;
}
.card-input.short {
  width: 90px;
}
.card-input.bordered {
  border: 1.5px solid #334155;
  border-radius: 0.7rem;
  outline: none;
  transition: border-color 0.15s;
}
.card-input.bordered:focus {
  border-color: #60a5fa;
}
.action-btn {
  margin-top: 1.2rem;
  font-size: 1.1rem;
  padding: 0.9rem 0;
  border-radius: 0.7rem;
}
</style>
