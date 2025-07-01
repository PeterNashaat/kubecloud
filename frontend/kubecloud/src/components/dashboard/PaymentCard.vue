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
        <label class="balance-label" for="current-balance-input">Current Balance:</label>
        <input
          id="current-balance-input"
          class="balance-value balance-input"
          type="text"
          :value="`$ ${user?.credited_balance ?? 0}`"
          readonly
          tabindex="-1"
        />
      </div>
      <div class="amount-section">
        <label class="section-label">Amount</label>
        <div class="amount-options">
          <button
            v-for="preset in presets"
            :key="preset"
            type="button"
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
          <div ref="cardElementRef" class="card-input" style="width:100%"></div>
        </div>
      </div>
      <div class="charge-section">
        <v-btn
          variant="outlined"
          class="action-btn"
          color="primary"
          :loading="loading"
          :disabled="loading"
          @click="chargeBalance"
          prepend-icon="mdi-credit-card-plus"
        >
          Charge Balance
        </v-btn>
      </div>
      <div v-if="successMessage" class="success-message">{{ successMessage }}</div>
      <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import { userService } from '../../utils/userService'
import { loadStripe } from '@stripe/stripe-js'
import type { Stripe, StripeElements } from '@stripe/stripe-js'

const { user } = storeToRefs(useUserStore())

const presets = [5, 10, 20, 50]
const amount = ref<number | 'custom'>(5)
const customAmount = ref<number | null>(null)
const loading = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

// Stripe integration
const stripe = ref<Stripe | null>(null)
const elements = ref<StripeElements | null>(null)
const cardElement = ref<any>(null)
const cardElementRef = ref<HTMLDivElement | null>(null)

// Use only import.meta.env for publishable key
// Make sure VITE_STRIPE_PUBLISHABLE_KEY is set in your .env file
const STRIPE_PUBLISHABLE_KEY = import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY

onMounted(async () => {
  if (!STRIPE_PUBLISHABLE_KEY) {
    console.error('VITE_STRIPE_PUBLISHABLE_KEY is not set in environment variables')
    errorMessage.value = 'Stripe configuration is missing. Please contact support.'
    return
  }
  
  try {
    stripe.value = await loadStripe(STRIPE_PUBLISHABLE_KEY)
    if (stripe.value) {
      elements.value = stripe.value.elements()
      if (cardElementRef.value) {
        cardElement.value = elements.value.create('card', {
          style: {
            base: {
              fontSize: '1rem',
              color: '#CBD5E1',
              fontFamily: 'inherit',
              backgroundColor: '#232f47',
              border: '1.5px solid #334155',
              borderRadius: '0.75rem',
              padding: '0.5rem 1rem',
              width: '100%',
              '::placeholder': {
                color: '#64748b',
              },
            },
            focus: {
              color: '#38BDF8',
              backgroundColor: '#232f47',
              border: '1.5px solid #38BDF8',
            },
            invalid: {
              color: '#ef4444',
              iconColor: '#ef4444',
              backgroundColor: '#232f47',
              border: '1.5px solid #ef4444',
            },
            complete: {
              color: '#fff',
              backgroundColor: '#232f47',
              border: '1.5px solid #10B981',
            },
          },
          hidePostalCode: true
        })
        cardElement.value.mount(cardElementRef.value)
      }
    }
  } catch (error) {
    console.error('Failed to initialize Stripe:', error)
    errorMessage.value = 'Failed to initialize payment system. Please refresh the page.'
  }
})

function selectAmount(val: number | 'custom') {
  amount.value = val
  if (val !== 'custom') customAmount.value = null
}

async function chargeBalance() {
  loading.value = true
  successMessage.value = ''
  errorMessage.value = ''
  const selectedAmount = getSelectedAmount()
  if (!selectedAmount) {
    errorMessage.value = 'Please select an amount.'
    loading.value = false
    return
  }
  if (!stripe.value || !elements.value) {
    errorMessage.value = 'Stripe is not loaded.'
    loading.value = false
    return
  }
  try {
    // 1. Create PaymentIntent on backend
    // @ts-ignore
    const intentRes = await userService.createPaymentIntent({ amount: Number(selectedAmount) })
    const clientSecret = intentRes.clientSecret
    if (!clientSecret) throw new Error('Failed to get payment intent.')

    // 2. Confirm card payment
    const { error, paymentIntent } = await stripe.value.confirmCardPayment(clientSecret, {
      payment_method: {
        card: cardElement.value,
      },
    })
    if (error) {
      errorMessage.value = error.message || 'Payment failed.'
      loading.value = false
      return
    }
    if (paymentIntent && paymentIntent.status === 'succeeded') {
      // 3. Notify backend to credit balance
      await userService.chargeBalance({
        card_type: 'stripe',
        payment_method_id: paymentIntent.payment_method as string,
        amount: Number(selectedAmount)
      })
      successMessage.value = 'Balance charged successfully!'
    } else {
      errorMessage.value = 'Payment not successful.'
    }
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
  width: 100%;
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
  background: transparent;
  border: none;
  outline: none;
  pointer-events: none;
  box-shadow: none;
  padding: 0;
  margin-left: 0.5rem;
  width: auto;
  min-width: 0;
  max-width: 7ch;
  height: 1.8rem;
  line-height: 1.2;
  display: inline-block;
  vertical-align: middle;
}
/* Style for readonly input */
.balance-input[readonly] {
  background: transparent !important;
  border: none !important;
  color: var(--color-success, #10B981) !important;
  font-weight: 700;
  font-size: 1.2rem;
  outline: none !important;
  box-shadow: none !important;
  cursor: default;
  width: auto;
  min-width: 0;
  max-width: 7ch;
  height: 1.8rem;
  padding: 0;
  margin: 0 0.2rem;
  text-align: left;
  display: inline-block;
  vertical-align: middle;
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
  transition: background 0.15s, border-color 0.15s, color 0.15s, box-shadow 0.12s, transform 0.12s;
  outline: none;
}
.amount-btn:active {
  transform: scale(0.96);
  box-shadow: 0 2px 8px 0 rgba(56, 189, 248, 0.15);
  border-color: var(--color-primary, #38BDF8);
}
.amount-btn.selected {
  background: rgba(56, 189, 248, 0.12);
  border-color: var(--color-primary, #38BDF8);
  color: var(--color-primary, #38BDF8);
  box-shadow: 0 2px 8px 0 rgba(56, 189, 248, 0.18);
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
  transition: border-color 0.15s, box-shadow 0.15s, color 0.15s;
}
.amount-input.selected,
.amount-input:focus {
  border-color: var(--color-primary, #38BDF8);
  box-shadow: 0 2px 8px 0 rgba(56, 189, 248, 0.18);
  color: var(--color-primary, #38BDF8);
}
.card-details-section {
  margin-bottom: 0;
}
.card-details-fields {
  margin-top: 0.5rem;
  width: 100%;
  display: block;
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
  width: 100%;
  min-width: 350px;
  max-width: 480px;
  display: block;
  box-sizing: border-box;
  min-height: 48px;
  height: 48px;
}
.card-input.short {
  width: 5.5rem;
}
.charge-section {
  margin-top: 0.5rem;
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
