import { ref, computed } from 'vue'
import { userService } from '@/utils/userService'
import { api } from '@/utils/api'

// Interface for rented nodes (matches the grid proxy structure)
export interface RentedNode {
  id: number
  nodeId?: number
  farmId?: number
  farmName?: string
  twinId?: number
  name?: string
  location?: string
  country?: string
  city?: string
  gridVersion?: number
  uptime?: number
  created?: number
  updatedAt?: number
  total_resources?: {
    cru: number
    sru: number
    hru: number
    mru: number
  }
  used_resources?: {
    cru: number
    sru: number
    hru: number
    mru: number
  }
  resources?: {
    cpu: number
    memory: number
    storage: number
    sru?: number
    hru?: number
    mru?: number
  }
  gpu?: string
  gpus?: any[]
  num_gpu?: number
  price?: number
  price_usd?: number
  status?: string
  healthy?: boolean
  rentable?: boolean
  rented?: boolean
  rentContractId?: number
  rentedByTwinId?: number
  certificationType?: string
  dedicated?: boolean
  inDedicatedFarm?: boolean
  features?: string[]
}

export function useNodeManagement() {
  const rentedNodes = ref<RentedNode[]>([])
  const total = ref<number>(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Fetch user's rented nodes
  async function fetchRentedNodes() {
    loading.value = true
    error.value = null
    try {
      const response = await userService.listReservedNodes()
      const responseData = response.data as any
      
      if (responseData.data?.nodes) {
        // Map the grid proxy node structure to our frontend structure
        rentedNodes.value = responseData.data.nodes.map((node: any) => ({
          id: node.nodeId || node.id,
          nodeId: node.nodeId,
          farmId: node.farmId,
          farmName: node.farmName,
          twinId: node.twinId,
          name: node.name || `Node #${node.nodeId || node.id}`,
          location: node.location || node.city || node.country || 'Unknown',
          country: node.country,
          city: node.city,
          gridVersion: node.gridVersion,
          uptime: node.uptime,
          created: node.created,
          updatedAt: node.updatedAt,
          total_resources: node.total_resources,
          used_resources: node.used_resources,
          resources: {
            cpu: node.total_resources?.cru || 0,
            memory: Math.round((node.total_resources?.mru || 0) / (1024 * 1024 * 1024)), // Convert to GB
            storage: Math.round((node.total_resources?.sru || 0) / (1024 * 1024 * 1024)), // Convert to GB
            sru: node.total_resources?.sru,
            hru: node.total_resources?.hru,
            mru: node.total_resources?.mru,
          },
          gpu: node.gpu,
          gpus: node.gpus,
          num_gpu: node.num_gpu,
          price: node.price_usd,
          price_usd: node.price_usd,
          status: node.status,
          healthy: node.healthy,
          rentable: node.rentable,
          rented: node.rented,
          rentContractId: node.rentContractId,
          rentedByTwinId: node.rentedByTwinId,
          certificationType: node.certificationType,
          dedicated: node.dedicated,
          inDedicatedFarm: node.inDedicatedFarm,
          features: node.features,
        }))
        total.value = responseData.data.total || 0
      } else {
        rentedNodes.value = []
        total.value = 0
      }
    } catch (err: any) {
      console.error('Failed to fetch rented nodes:', err)
      error.value = err?.message || 'Failed to fetch rented nodes'
      rentedNodes.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  // Reserve a node
  async function reserveNode(nodeId: number) {
    try {
      const response = await userService.reserveNode(nodeId)
      // Refresh the rented nodes list after successful reservation
      await fetchRentedNodes()
      return response
    } catch (err: any) {
      console.error('Failed to reserve node:', err)
      throw err
    }
  }

  // Unreserve a node
  async function unreserveNode(contractId: string) {
    try {
      const response = await userService.unreserveNode(contractId)
      // Refresh the rented nodes list after successful unreservation
      await fetchRentedNodes()
      return response
    } catch (err: any) {
      console.error('Failed to unreserve node:', err)
      throw err
    }
  }

  // Calculate total monthly cost of rented nodes
  const totalMonthlyCost = computed(() => {
    return rentedNodes.value
      .filter(node => typeof node.price === 'number')
      .reduce((sum, node) => sum + (node.price || 0), 0)
  })

  // Get nodes by status
  const healthyNodes = computed(() => 
    rentedNodes.value.filter(node => node.healthy)
  )

  const unhealthyNodes = computed(() => 
    rentedNodes.value.filter(node => !node.healthy)
  )

  return {
    rentedNodes,
    total,
    loading,
    error,
    fetchRentedNodes,
    reserveNode,
    unreserveNode,
    totalMonthlyCost,
    healthyNodes,
    unhealthyNodes
  }
} 