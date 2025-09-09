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
 * Format storage size in GB with appropriate units (GB, TB)
 * @param gb - Size in GB
 * @param decimals - Number of decimal places (default: 1)
 * @returns Formatted string with unit
 */
export function formatStorageSize(gb: number, decimals: number = 1): string {
  if (gb === 0) return '0 GB'
  let formatted = gb.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: decimals })
  formatted = formatted.replace(/\.0+$/, '')
  return formatted + ' GB'
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
      value: formatStorageSize(stats.ssd),
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

