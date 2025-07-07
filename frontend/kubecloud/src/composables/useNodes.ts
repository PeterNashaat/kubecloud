import { ref, computed } from 'vue'
import { userService } from '../utils/userService'
import type { RawNode } from '../types/rawNode'
import type { NormalizedNode } from '../types/normalizedNode'
import { normalizeNode } from '../utils/nodeNormalizer'

export interface NodeFilters {
  country?: string
  city?: string
  farm_id?: number
  free_hru?: number
  free_mru?: number
  free_sru?: number
  free_cru?: number
  status?: string
  healthy?: boolean
  rentable?: boolean
  dedicated?: boolean
  certification_type?: string
  grid_version?: number
}

export function useNodes() {
  const nodes = ref<RawNode[]>([])
  const total = ref<number>(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const filters = ref<NodeFilters>({})

  // Fetch all available nodes
  async function fetchNodes(nodeFilters?: NodeFilters) {
    loading.value = true
    error.value = null
    try {
      const response = await userService.listNodes(nodeFilters || filters.value)
      const responseData = response.data as any
      if (responseData.data?.nodes) {
        nodes.value = responseData.data.nodes
        total.value = responseData.data.total || 0
      } else {
        nodes.value = []
        total.value = 0
      }
    } catch (err: any) {
      console.error('Failed to fetch nodes:', err)
      error.value = err?.message || 'Failed to fetch nodes'
      nodes.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  // Update filters and refetch
  async function updateFilters(newFilters: NodeFilters) {
    filters.value = { ...filters.value, ...newFilters }
    await fetchNodes()
  }

  // Clear all filters
  async function clearFilters() {
    filters.value = {}
    await fetchNodes()
  }

  // Normalized nodes for UI
  const normalizedNodes = computed<NormalizedNode[]>(() => nodes.value.map(normalizeNode))

  return {
    nodes, // RawNode[]
    normalizedNodes, // NormalizedNode[]
    total,
    loading,
    error,
    filters,
    fetchNodes,
    updateFilters,
    clearFilters
  }
} 