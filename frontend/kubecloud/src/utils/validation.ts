export interface ValidationRule {
  required?: boolean
  minLength?: number
  maxLength?: number
  pattern?: RegExp
  email?: boolean
  url?: boolean
  custom?: (value: any) => boolean | string
}

export interface ValidationResult {
  isValid: boolean
  errors: string[]
}

export interface FieldValidation {
  value: any
  rules: ValidationRule
  fieldName?: string
}

export class ValidationError extends Error {
  constructor(message: string, public field?: string) {
    super(message)
    this.name = 'ValidationError'
  }
}

// Common validation patterns
export const PATTERNS = {
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  URL: /^https?:\/\/.+/,
  PHONE: /^\+?[\d\s\-\(\)]+$/,
  PASSWORD: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/,
  ALPHANUMERIC: /^[a-zA-Z0-9]+$/,
  HEX_COLOR: /^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$/,
  IP_ADDRESS: /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
}

// Validation functions
export const validateField = (field: FieldValidation): ValidationResult => {
  const { value, rules, fieldName = 'Field' } = field
  const errors: string[] = []

  // Required validation
  if (rules.required && (!value || (typeof value === 'string' && value.trim() === ''))) {
    errors.push(`${fieldName} is required`)
  }

  // Skip other validations if value is empty and not required
  if (!value || (typeof value === 'string' && value.trim() === '')) {
    return { isValid: errors.length === 0, errors }
  }

  // Type-specific validations
  if (typeof value === 'string') {
    // Length validations
    if (rules.minLength && value.length < rules.minLength) {
      errors.push(`${fieldName} must be at least ${rules.minLength} characters`)
    }
    if (rules.maxLength && value.length > rules.maxLength) {
      errors.push(`${fieldName} must be no more than ${rules.maxLength} characters`)
    }

    // Pattern validations
    if (rules.pattern && !rules.pattern.test(value)) {
      errors.push(`${fieldName} format is invalid`)
    }

    // Email validation
    if (rules.email && !PATTERNS.EMAIL.test(value)) {
      errors.push(`${fieldName} must be a valid email address`)
    }

    // URL validation
    if (rules.url && !PATTERNS.URL.test(value)) {
      errors.push(`${fieldName} must be a valid URL`)
    }
  }

  // Number validations
  if (typeof value === 'number') {
    if (rules.minLength && value < rules.minLength) {
      errors.push(`${fieldName} must be at least ${rules.minLength}`)
    }
    if (rules.maxLength && value > rules.maxLength) {
      errors.push(`${fieldName} must be no more than ${rules.maxLength}`)
    }
  }

  // Custom validation
  if (rules.custom) {
    const customResult = rules.custom(value)
    if (customResult !== true) {
      errors.push(typeof customResult === 'string' ? customResult : `${fieldName} is invalid`)
    }
  }

  return { isValid: errors.length === 0, errors }
}

export const validateForm = (fields: Record<string, FieldValidation>): ValidationResult => {
  const allErrors: string[] = []
  let isValid = true

  for (const [fieldName, field] of Object.entries(fields)) {
    const result = validateField({
      ...field,
      fieldName: field.fieldName || fieldName
    })

    if (!result.isValid) {
      isValid = false
      allErrors.push(...result.errors)
    }
  }

  return { isValid, errors: allErrors }
}

// Common validation rules
export const VALIDATION_RULES = {
  REQUIRED: { required: true },
  EMAIL: { required: true, email: true },
  PASSWORD: {
    required: true,
    minLength: 8,
    pattern: PATTERNS.PASSWORD,
    custom: (value: string) => {
      if (!PATTERNS.PASSWORD.test(value)) {
        return 'Password must contain at least 8 characters, including uppercase, lowercase, number, and special character (@$!%*?&)'
      }
      return true
    }
  },
  URL: { required: true, url: true },
  PHONE: { required: true, pattern: PATTERNS.PHONE },
  ALPHANUMERIC: { required: true, pattern: PATTERNS.ALPHANUMERIC },
  HEX_COLOR: { pattern: PATTERNS.HEX_COLOR },
  IP_ADDRESS: { pattern: PATTERNS.IP_ADDRESS },
  CREDIT_AMOUNT: {
    required: true,
    custom: (value: number) => {
      if (typeof value !== 'number' || isNaN(value)) {
        return 'Amount must be a valid number'
      }
      if (value <= 0) {
        return 'Amount must be greater than 0'
      }
      if (value > 10000) {
        return 'Amount cannot exceed $10,000'
      }
      if (!/^\d+(\.\d{1,2})?$/.test(value.toString())) {
        return 'Amount can have at most 2 decimal places'
      }
      return true
    }
  },
  CREDIT_MEMO: {
    required: true,
    minLength: 3,
    maxLength: 255,
    custom: (value: string) => {
      if (typeof value !== 'string') {
        return 'Memo must be a string'
      }
      const trimmed = value.trim()
      if (trimmed.length < 3) {
        return 'Memo must be at least 3 characters long'
      }
      if (trimmed.length > 255) {
        return 'Memo cannot exceed 255 characters'
      }
      if (!/^[a-zA-Z0-9\s\-_.,!?()]+$/.test(trimmed)) {
        return 'Memo contains invalid characters. Only letters, numbers, spaces, and basic punctuation are allowed'
      }
      return true
    }
  }
}

// Utility functions
export const sanitizeInput = (input: string): string => {
  return input.trim().replace(/[<>]/g, '')
}

// Credit operation validation functions
export const validateCreditAmount = (amount: number): { isValid: boolean; error: string } => {
  const validation = validateField({
    value: amount,
    rules: VALIDATION_RULES.CREDIT_AMOUNT,
    fieldName: 'Amount'
  })

  return {
    isValid: validation.isValid,
    error: validation.errors[0] || ''
  }
}

export const validateCreditMemo = (memo: string): { isValid: boolean; error: string } => {
  const validation = validateField({
    value: memo,
    rules: VALIDATION_RULES.CREDIT_MEMO,
    fieldName: 'Memo'
  })

  return {
    isValid: validation.isValid,
    error: validation.errors[0] || ''
  }
}

export const validateCreditForm = (amount: number, memo: string): { isValid: boolean; errors: { amount: string; memo: string } } => {
  const amountValidation = validateCreditAmount(amount)
  const memoValidation = validateCreditMemo(memo)

  return {
    isValid: amountValidation.isValid && memoValidation.isValid,
    errors: {
      amount: amountValidation.error,
      memo: memoValidation.error
    }
  }
}

// Additional validation utilities for auth operations
export const validateVerificationCode = (code: string): { isValid: boolean; error: string } => {
  if (!code.trim()) {
    return { isValid: false, error: 'Verification code is required' }
  }
  if (code.length < 4 || code.length > 6) {
    return { isValid: false, error: 'Verification code must be 4-6 digits' }
  }
  if (!/^\d+$/.test(code)) {
    return { isValid: false, error: 'Verification code must contain only numbers' }
  }
  return { isValid: true, error: '' }
}

export const validateEmail = (email: string): { isValid: boolean; error: string } => {
  if (!email.trim()) {
    return { isValid: false, error: 'Email is required' }
  }
  if (!PATTERNS.EMAIL.test(email.trim())) {
    return { isValid: false, error: 'Please enter a valid email address' }
  }
  return { isValid: true, error: '' }
}

export const validatePasswordStrength = (password: string): { isValid: boolean; error: string; strength: 'weak' | 'medium' | 'strong' } => {
  if (!password) {
    return { isValid: false, error: 'Password is required', strength: 'weak' }
  }

  const validation = validateField({
    value: password,
    rules: VALIDATION_RULES.PASSWORD,
    fieldName: 'Password'
  })

  if (!validation.isValid) {
    return { isValid: false, error: validation.errors[0], strength: 'weak' }
  }

  // Determine strength
  let strength: 'weak' | 'medium' | 'strong' = 'medium'
  if (password.length >= 12 && /[A-Z].*[A-Z]/.test(password) && /[0-9].*[0-9]/.test(password)) {
    strength = 'strong'
  } else if (password.length < 10) {
    strength = 'weak'
  }

  return { isValid: true, error: '', strength }
}

export const validateConfirmPassword = (password: string, confirmPassword: string): { isValid: boolean; error: string } => {
  if (!confirmPassword) {
    return { isValid: false, error: 'Please confirm your password' }
  }
  if (password !== confirmPassword) {
    return { isValid: false, error: 'Passwords do not match' }
  }
  return { isValid: true, error: '' }
}

export const formatValidationErrors = (errors: string[]): string => {
  return errors.join('. ')
}

export const createValidationRule = (rule: ValidationRule): ValidationRule => {
  return rule
}

// Async validation support
export const validateAsync = async (
  field: FieldValidation,
  asyncValidator?: (value: any) => Promise<boolean | string>
): Promise<ValidationResult> => {
  const syncResult = validateField(field)

  if (!syncResult.isValid || !asyncValidator) {
    return syncResult
  }

  try {
    const asyncResult = await asyncValidator(field.value)
    if (asyncResult !== true) {
      return {
        isValid: false,
        errors: [...syncResult.errors, typeof asyncResult === 'string' ? asyncResult : 'Validation failed']
      }
    }
  } catch (error) {
    return {
      isValid: false,
      errors: [...syncResult.errors, 'Validation error occurred']
    }
  }

  return syncResult
}





export function required(msg: string) {
  return (value: string) => {
    if (!value) {
      return msg;
    }
  };
}
export function min(msg: string, min: number) {
  return (value: number) => {
    if (value < min) {
      return msg ;
    }
  };
}

export function max(msg: string, max: number) {
  return (value: number) => {
    if (+value > max) {
      return msg;
    }
  };
}

export function isAlphanumeric(msg: string) {
  return (value: string) => {
    if (!/^[a-zA-Z0-9]*$/.test(value)) {
      return msg;
    }
  };
}
