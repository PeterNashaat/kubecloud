import { ref, computed } from 'vue'
import { userService, type Node } from '../utils/userService'

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
  const nodes = ref<Node[]>([])
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

  // Get nodes by country
  const nodesByCountry = computed(() => {
    const countries = new Map<string, Node[]>()
    nodes.value.forEach(node => {
      const country = node.country || 'Unknown'
      if (!countries.has(country)) {
        countries.set(country, [])
      }
      countries.get(country)!.push(node)
    })
    return countries
  })

  // Get healthy nodes
  const healthyNodes = computed(() => 
    nodes.value.filter(node => node.healthy)
  )

  // Get rentable nodes
  const rentableNodes = computed(() => 
    nodes.value.filter(node => node.rentable)
  )

  // Get dedicated nodes
  const dedicatedNodes = computed(() => 
    nodes.value.filter(node => node.dedicated)
  )

  // Get nodes with GPU (filter by certification type)
  const gpuNodes = computed(() => 
    nodes.value.filter(node => node.certification_type === 'Certified' || node.certification_type === 'Gold')
  )

  // Get average resources across all nodes
  const averageResources = computed(() => {
    if (nodes.value.length === 0) return null
    
    const total = nodes.value.reduce((acc, node) => ({
      cru: acc.cru + (node.resources?.cru || 0),
      mru: acc.mru + (node.resources?.mru || 0),
      sru: acc.sru + (node.resources?.sru || 0),
      hru: acc.hru + (node.resources?.hru || 0)
    }), { cru: 0, mru: 0, sru: 0, hru: 0 })

    return {
      cru: Math.round(total.cru / nodes.value.length),
      mru: Math.round(total.mru / nodes.value.length / (1024 * 1024 * 1024)), // Convert to GB
      sru: Math.round(total.sru / nodes.value.length / (1024 * 1024 * 1024)), // Convert to GB
      hru: Math.round(total.hru / nodes.value.length / (1024 * 1024 * 1024))  // Convert to GB
    }
  })

  return {
    nodes,
    total,
    loading,
    error,
    filters,
    fetchNodes,
    updateFilters,
    clearFilters,
    nodesByCountry,
    healthyNodes,
    rentableNodes,
    dedicatedNodes,
    gpuNodes,
    averageResources
  }
} 