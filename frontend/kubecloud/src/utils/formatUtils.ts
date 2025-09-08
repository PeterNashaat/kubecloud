/**
 * Utility functions for formatting data display
 */

/**
 * Format large numbers with appropriate units (K, M, B, T)
 * @param value - The number to format
 * @param decimals - Number of decimal places (default: 1)
 * @returns Formatted string with unit
 */
export function formatLargeNumber(value: number, decimals: number = 1): string {
  if (value === 0) return '0'
  
  const units = ['', 'K', 'M', 'B', 'T']
  const k = 1000
  const dm = decimals < 0 ? 0 : decimals
  
  const i = Math.floor(Math.log(Math.abs(value)) / Math.log(k))
  const unitIndex = Math.min(i, units.length - 1)
  
  if (unitIndex === 0) {
    return value.toString()
  }
  
  const formattedValue = (value / Math.pow(k, unitIndex)).toFixed(dm)
  // Remove trailing zeros and decimal point if not needed
  const cleanValue = parseFloat(formattedValue).toString()
  
  return cleanValue + units[unitIndex]
}

/**
 * Format storage size with appropriate units (B, KB, MB, GB, TB, PB)
 * @param bytes - Size in bytes
 * @param decimals - Number of decimal places (default: 1)
 * @returns Formatted string with unit
 */
export function formatStorageSize(bytes: number, decimals: number = 1): string {
  if (bytes === 0) return '0 B'
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  
  const i = Math.floor(Math.log(Math.abs(bytes)) / Math.log(k))
  const unitIndex = Math.min(i, units.length - 1)
  
  if (unitIndex === 0) {
    return bytes + ' B'
  }
  
  const formattedValue = (bytes / Math.pow(k, unitIndex)).toFixed(dm)
  // Remove trailing zeros and decimal point if not needed
  const cleanValue = parseFloat(formattedValue).toString()
  
  return cleanValue + ' ' + units[unitIndex]
}

/**
 * Format currency values
 * @param value - The currency value
 * @param currency - Currency symbol (default: '$')
 * @param decimals - Number of decimal places (default: 2)
 * @returns Formatted currency string
 */
export function formatCurrency(value: number, currency: string = '$', decimals: number = 2): string {
  return currency + value.toFixed(decimals)
}

/**
 * Format percentage values
 * @param value - The percentage value (0-100)
 * @param decimals - Number of decimal places (default: 1)
 * @returns Formatted percentage string
 */
export function formatPercentage(value: number, decimals: number = 1): string {
  return value.toFixed(decimals) + '%'
}

/**
 * Format statistics data for display cards
 * @param stats - Raw statistics object
 * @returns Formatted statistics for display
 */
export interface FormattedStat {
  label: string
  value: string
  rawValue: number
}

export function formatStatsForCards(stats: {
  ssd: number
  up_nodes: number
  countries: number
  cores: number
  total_users?: number
  total_clusters?: number
}): FormattedStat[] {
  return [
    {
      label: 'SSD Storage',
      value: formatStorageSize(stats.ssd * 1024 * 1024 * 1024), // Convert GB to bytes for proper formatting
      rawValue: stats.ssd
    },
    {
      label: 'Active Nodes',
      value: formatLargeNumber(stats.up_nodes, 0),
      rawValue: stats.up_nodes
    },
    {
      label: 'Countries',
      value: stats.countries.toString(),
      rawValue: stats.countries
    },
    {
      label: 'CPU Cores',
      value: formatLargeNumber(stats.cores, 0),
      rawValue: stats.cores
    }
  ]
}

/**
 * Format uptime in a human-readable format
 * @param hours - Uptime in hours
 * @returns Formatted uptime string
 */
export function formatUptime(hours: number): string {
  if (hours < 24) {
    return `${Math.round(hours)}h`
  } else if (hours < 24 * 7) {
    const days = Math.floor(hours / 24)
    const remainingHours = Math.round(hours % 24)
    return remainingHours > 0 ? `${days}d ${remainingHours}h` : `${days}d`
  } else if (hours < 24 * 30) {
    const weeks = Math.floor(hours / (24 * 7))
    const remainingDays = Math.floor((hours % (24 * 7)) / 24)
    return remainingDays > 0 ? `${weeks}w ${remainingDays}d` : `${weeks}w`
  } else {
    const months = Math.floor(hours / (24 * 30))
    const remainingWeeks = Math.floor((hours % (24 * 30)) / (24 * 7))
    return remainingWeeks > 0 ? `${months}mo ${remainingWeeks}w` : `${months}mo`
  }
}
