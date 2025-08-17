<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <h3 class="dashboard-card-title">Manual Credit</h3>
      <p class="section-subtitle">Apply credits to user accounts</p>
    </div>
    
    <v-form ref="formRef" @submit.prevent="handleSubmit" class="credit-form">
      <div class="form-row">
        <v-text-field
          v-model.number="creditAmount"
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
          :rules="[rules.required, rules.creditAmount]"
          :disabled="isSubmitting"
          placeholder="Enter amount (e.g., 25.50)"
        />
        <v-text-field
          v-model="creditReason"
          label="Reason/Memo"
          prepend-inner-icon="mdi-note-text"
          variant="outlined"
          density="comfortable"
          minlength="3"
          maxlength="255"
          required
          class="form-field mb-3"
          :rules="[rules.required, rules.creditMemo]"
          :disabled="isSubmitting"
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
        {{ isSubmitting ? 'Applying Credit...' : 'Apply Credit' }}
      </v-btn>
      
      <!-- Success Message -->
      <v-alert
        v-if="showSuccess"
        type="success"
        variant="tonal"
        class="mt-4"
        closable
      >
        <template #prepend>
          <v-icon icon="mdi-check-circle" />
        </template>
        Credit applied successfully!
      </v-alert>
    </v-form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { validateCreditAmount, validateCreditMemo } from '../utils/validation'

const emit = defineEmits(['applyManualCredit'])

const isSubmitting = ref(false)
const showSuccess = ref(false)
const creditAmount = ref(0)
const creditReason = ref('')

// Validation rules using existing validation utilities
const rules = {
  required: (value: any) => !!value || 'This field is required',
  creditAmount: (value: number) => {
    const validation = validateCreditAmount(value)
    return validation.isValid || validation.error
  },
  creditMemo: (value: string) => {
    const validation = validateCreditMemo(value)
    return validation.isValid || validation.error
  }
}

const isFormValid = computed(() => {
  const amountValidation = validateCreditAmount(creditAmount.value)
  const memoValidation = validateCreditMemo(creditReason.value)
  return amountValidation.isValid && memoValidation.isValid
})

const handleSubmit = async () => {
  if (!isFormValid.value) return

  try {
    isSubmitting.value = true
    await emit('applyManualCredit', {
      amount: creditAmount.value,
      reason: creditReason.value.trim()
    })
    
    // Show success message
    showSuccess.value = true
    
    // Reset form after successful submission
    creditAmount.value = 0
    creditReason.value = ''
    
    // Hide success message after 3 seconds
    setTimeout(() => {
      showSuccess.value = false
    }, 3000)
    
  } catch (error) {
    console.error('Credit application error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>
