import { formatDistanceToNow } from 'date-fns'

/**
 * Get notification icon for different types
 * @param type - Notification type
 * @returns Icon name
 */
export function getNotificationIcon(type: string): string {
  switch (type) {
    case 'deployment_update': return 'mdi-rocket-launch'
    case 'task_update': return 'mdi-cog'
    case 'connected': return 'mdi-link'
    case 'error': return 'mdi-alert-circle'
    case 'success': return 'mdi-check-circle'
    case 'warning': return 'mdi-alert'
    case 'info': return 'mdi-information'
    default: return 'mdi-bell'
  }
}

/**
 * Get notification color for different types
 * @param type - Notification type
 * @returns Color name
 */
export function getNotificationColor(type: string): string {
  switch (type) {
    case 'deployment_update': return 'success'
    case 'task_update': return 'info'
    case 'connected': return 'primary'
    case 'error': return 'error'
    case 'success': return 'success'
    case 'warning': return 'warning'
    case 'info': return 'info'
    default: return 'grey'
  }
}

/**
 * Get toast notification icon for different types
 * @param type - Notification type
 * @returns Icon name
 */
export function getToastIcon(type: string): string {
  switch (type) {
    case 'success': return 'mdi-check-circle'
    case 'error': return 'mdi-alert-circle'
    case 'warning': return 'mdi-alert'
    case 'info': return 'mdi-information'
    default: return 'mdi-information'
  }
}

/**
 * Get toast notification color for different types
 * @param type - Notification type
 * @returns Color hex value
 */
export function getToastColor(type: string): string {
  switch (type) {
    case 'success': return '#10B981'
    case 'error': return '#EF4444'
    case 'warning': return '#F59E0B'
    case 'info': return '#60a5fa'
    default: return '#60a5fa'
  }
}

/**
 * Format notification timestamp to relative time
 * @param timestamp - ISO timestamp string
 * @returns Formatted relative time string
 */
export function formatNotificationTime(timestamp: string): string {
  try {
    return formatDistanceToNow(new Date(timestamp), { addSuffix: true })
  } catch {
    return 'Unknown time'
  }
}
