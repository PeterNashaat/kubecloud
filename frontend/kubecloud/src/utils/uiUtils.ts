/**
 * Truncate text to a specified length with ellipsis
 * @param text - Text to truncate
 * @param maxLength - Maximum length
 * @returns Truncated text
 */
export function truncateText(text: string, maxLength: number = 50): string {
  if (!text || text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

/**
 * Format bytes to human readable format
 * @param bytes - Number of bytes
 * @param decimals - Number of decimal places
 * @returns Formatted string
 */
export function formatBytes(bytes: number, decimals: number = 2): string {
  if (bytes === 0) return '0 Bytes'
  
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

/**
 * Format currency
 * @param amount - Amount to format
 * @param currency - Currency code
 * @returns Formatted currency string
 */
export function formatCurrency(amount: number, currency: string = 'USD'): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency
  }).format(amount)
}

/**
 * Get status color for different states
 * @param status - Status string
 * @returns Color string
 */
export function getStatusColor(status: string): string {
  const statusColors: Record<string, string> = {
    'active': 'success',
    'running': 'success',
    'healthy': 'success',
    'completed': 'success',
    'pending': 'warning',
    'processing': 'warning',
    'stopped': 'error',
    'failed': 'error',
    'error': 'error',
    'inactive': 'grey',
    'unknown': 'grey'
  }
  
  return statusColors[status.toLowerCase()] || 'grey'
}

/**
 * Get status icon for different states
 * @param status - Status string
 * @returns Icon name
 */
export function getStatusIcon(status: string): string {
  const statusIcons: Record<string, string> = {
    'active': 'mdi-check-circle',
    'running': 'mdi-play-circle',
    'healthy': 'mdi-heart',
    'completed': 'mdi-check-circle',
    'pending': 'mdi-clock',
    'processing': 'mdi-sync',
    'stopped': 'mdi-stop-circle',
    'failed': 'mdi-alert-circle',
    'error': 'mdi-alert-circle',
    'inactive': 'mdi-pause-circle',
    'unknown': 'mdi-help-circle'
  }
  
  return statusIcons[status.toLowerCase()] || 'mdi-help-circle'
}

/**
 * Generate a unique ID
 * @returns Unique ID string
 */
export function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2)
}

/**
 * Debounce function execution
 * @param func - Function to debounce
 * @param wait - Wait time in milliseconds
 * @returns Debounced function
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout>
  
  return (...args: Parameters<T>) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

/**
 * Throttle function execution
 * @param func - Function to throttle
 * @param limit - Time limit in milliseconds
 * @returns Throttled function
 */
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean
  
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => inThrottle = false, limit)
    }
  }
} 