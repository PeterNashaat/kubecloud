import { reactive, computed } from 'vue'
import { validateForm, VALIDATION_RULES, type ValidationRule } from '../utils/validation'

export interface FormField {
  value: any
  rules: ValidationRule
  fieldName?: string
}

export interface FormValidationState {
  errors: Record<string, string>
  isValid: boolean
  touched: Record<string, boolean>
}

export function useFormValidation<T extends Record<string, FormField>>(
  fields: T,
  initialTouched: Record<string, boolean> = {}
) {
  const state = reactive<FormValidationState>({
    errors: {},
    isValid: false,
    touched: { ...initialTouched }
  })

  // Validate all fields
  const validateAll = () => {
    const result = validateForm(fields)
    state.isValid = result.isValid
    
    // Clear previous errors
    state.errors = {}
    
    // Map errors to field names
    result.errors.forEach(error => {
      for (const [fieldName, field] of Object.entries(fields)) {
        if (error.includes(field.fieldName || fieldName)) {
          state.errors[fieldName] = error
          break
        }
      }
    })
    
    return result.isValid
  }

  // Validate single field
  const validateField = (fieldName: keyof T) => {
    const field = fields[fieldName]
    if (!field) return true

    const result = validateForm({
      [fieldName as string]: {
        ...field,
        fieldName: field.fieldName || fieldName as string
      }
    })

    if (result.isValid) {
      delete state.errors[fieldName as string]
    } else {
      state.errors[fieldName as string] = result.errors[0] || 'Invalid field'
    }

    state.isValid = Object.keys(state.errors).length === 0
    return result.isValid
  }

  // Mark field as touched
  const touchField = (fieldName: keyof T) => {
    state.touched[fieldName as string] = true
  }

  // Clear all errors
  const clearErrors = () => {
    state.errors = {}
    state.isValid = false
  }

  // Reset form state
  const reset = () => {
    state.errors = {}
    state.isValid = false
    state.touched = { ...initialTouched }
  }

  // Get error for specific field
  const getFieldError = (fieldName: keyof T): string => {
    return state.errors[fieldName as string] || ''
  }

  // Check if field has error
  const hasFieldError = (fieldName: keyof T): boolean => {
    return !!state.errors[fieldName as string]
  }

  // Check if field is touched
  const isFieldTouched = (fieldName: keyof T): boolean => {
    return !!state.touched[fieldName as string]
  }

  return {
    // State
    errors: computed(() => state.errors),
    isValid: computed(() => state.isValid),
    touched: computed(() => state.touched),
    
    // Methods
    validateAll,
    validateField,
    touchField,
    clearErrors,
    reset,
    getFieldError,
    hasFieldError,
    isFieldTouched
  }
}

// Predefined validation rules for common use cases
export const FORM_VALIDATION_RULES = {
  ...VALIDATION_RULES,
  CLUSTER_NAME: { 
    required: true, 
    minLength: 3, 
    maxLength: 50,
    pattern: /^[a-zA-Z0-9_-]+$/
  },
  NODE_NAME: { 
    required: true, 
    minLength: 1, 
    maxLength: 30,
    pattern: /^[a-zA-Z0-9_-]+$/
  },
  CPU_COUNT: { 
    required: true,
    custom: (value: number) => value > 0 && value <= 64 ? true : 'CPU count must be between 1 and 64'
  },
  RAM_GB: { 
    required: true,
    custom: (value: number) => value > 0 && value <= 512 ? true : 'RAM must be between 1 and 512 GB'
  },
  STORAGE_GB: { 
    required: true,
    custom: (value: number) => value > 0 && value <= 10000 ? true : 'Storage must be between 1 and 10000 GB'
  }
} 