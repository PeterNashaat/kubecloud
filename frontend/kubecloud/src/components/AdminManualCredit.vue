<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <h3 class="dashboard-card-title">Manual Credit</h3>
      <p class="section-subtitle">Apply credits to user accounts</p>
    </div>
    <v-form ref="formRef" @submit.prevent="handleSubmit" class="credit-form">
      <div class="form-row">
        <v-text-field
          v-model.number="creditAmountLocal"
          label="Amount ($)"
          type="number"
          prepend-inner-icon="mdi-currency-usd"
          variant="outlined"
          min="0.01"
          max="10000"
          step="0.01"
          density="comfortable"
          required
          class="form-field mb-3"
          :error-messages="errors.amount"
          :disabled="isSubmitting"
          @blur="validateAmount"
          @input="clearAmountError"
          placeholder="Enter amount (e.g., 25.50)"
        />
        <v-text-field
          v-model="creditReasonLocal"
          label="Reason/Memo"
          prepend-inner-icon="mdi-note-text"
          variant="outlined"
          density="comfortable"
          minlength="3"
          maxlength="255"
          required
          class="form-field mb-3"
          :error-messages="errors.memo"
          :disabled="isSubmitting"
          @blur="validateMemo"
          @input="clearMemoError"
          placeholder="Enter reason for credit (e.g., Account adjustment)"
          counter="255"
        />
      </div>
      <v-btn
        type="submit"
        color="primary"
        variant="elevated"
        class="btn-primary"
        :disabled="!isFormValid || isSubmitting"
        :loading="isSubmitting"
      >
        <v-icon icon="mdi-cash-plus" class="mr-2"></v-icon>
        Apply Credit
      </v-btn>
    </v-form>
  </div>
</template>
<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { validateCreditForm, validateCreditAmount, validateCreditMemo } from '../utils/validation'

const props = defineProps({
  creditAmount: Number,
  creditReason: String,
})

const emit = defineEmits(['applyManualCredit', 'update:creditAmount', 'update:creditReason'])

const formRef = ref()
const isSubmitting = ref(false)
const creditAmountLocal = ref(props.creditAmount || 0)
const creditReasonLocal = ref(props.creditReason || '')

const errors = ref({
  amount: '',
  memo: ''
})

const isFormValid = computed(() => {
  return creditAmountLocal.value > 0 &&
         creditReasonLocal.value.trim().length >= 3 &&
         !errors.value.amount &&
         !errors.value.memo
})

const validateAmount = () => {
  const validation = validateCreditAmount(creditAmountLocal.value)
  errors.value.amount = validation.error
}

const validateMemo = () => {
  const validation = validateCreditMemo(creditReasonLocal.value)
  errors.value.memo = validation.error
}

const clearAmountError = () => {
  errors.value.amount = ''
}

const clearMemoError = () => {
  errors.value.memo = ''
}

const clearAllErrors = () => {
  errors.value.amount = ''
  errors.value.memo = ''
}

const handleSubmit = async () => {
  clearAllErrors()

  const validation = validateCreditForm(creditAmountLocal.value, creditReasonLocal.value)

  if (!validation.isValid) {
    errors.value = validation.errors
    return
  }

  try {
    isSubmitting.value = true
    emit('applyManualCredit')
  } catch (error) {
    console.error('Credit application error:', error)
  } finally {
    isSubmitting.value = false
  }
}

watch(() => props.creditAmount, val => {
  creditAmountLocal.value = val || 0
  clearAmountError()
})

watch(() => props.creditReason, val => {
  creditReasonLocal.value = val || ''
  clearMemoError()
})

watch(creditAmountLocal, val => {
  emit('update:creditAmount', val)
  if (val > 0) clearAmountError()
})

watch(creditReasonLocal, val => {
  emit('update:creditReason', val)
  if (val.trim().length >= 3) clearMemoError()
})
</script>
