import { ref, computed } from 'vue'
import { userService } from '@/utils/userService'
import { useNotificationStore } from '@/stores/notifications'

// Interface for rented nodes (matches the grid proxy structure)
export interface RentedNode {
  id: string;
  nodeId: number;
  farmId: number;
  farmName: string;
  twinId: number;
  country: string;
  gridVersion: number;
  city: string;
  uptime: number;
  created: number;
  farmingPolicyId: number;
  updatedAt: number;
  total_resources: {
    cru: number;
    sru: number;
    hru: number;
    mru: number;
  };
  used_resources: {
    cru: number;
    sru: number;
    hru: number;
    mru: number;
  };
  resources?: {
    cpu: number;
    memory: number;
    storage: number;
    sru?: number;
    hru?: number;
    mru?: number;
  };
  gpu?: boolean;
  gpus: any[];
  price_usd: number;
  farm_free_ips: number;
  features: string[];
  location: {
    country: string;
    city: string;
    longitude: number;
    latitude: number;
  };
  publicConfig: {
    domain: string;
    gw4: string;
    gw6: string;
    ipv4: string;
    ipv6: string;
  };
  status: string;
  certificationType: string;
  dedicated: boolean;
  inDedicatedFarm: boolean;
  rentContractId: number;
  rented: boolean;
  rentable: boolean;
  rentedByTwinId: number;
  serialNumber: string;
  power: {
    state: string;
    target: string;
  };
  num_gpu: number;
  extraFee: number;
  healthy: boolean;
  dmi: {
    bios: {
      vendor: string;
      version: string;
    };
    baseboard: {
      manufacturer: string;
      product_name: string;
    };
    processor: Array<{
      version: string;
      thread_count: string;
    }>;
    memory: Array<{
      manufacturer: string;
      type: string;
    }>;
  };
  speed: {
    upload: number;
    download: number;
  };
}

export function useNodeManagement() {
  const rentedNodes = ref<RentedNode[]>([])
  const total = ref<number>(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const reserveNodeLoading = ref(false)
  const notificationStore = useNotificationStore()

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
          nodeId: node.nodeId,
          farmId: node.farmId,
          farmName: node.farmName,
          twinId: node.twinId,
          country: node.country,
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
          gpus: node.gpus,
          num_gpu: node.num_gpu,
          price_usd: node.price_usd,
          status: node.status,
          healthy: node.healthy,
          rentContractId: node.rentContractId,
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
    reserveNodeLoading.value = true
    try {
      await userService.reserveNode(nodeId)
      // Refresh the rented nodes list after successful reservation
      await fetchRentedNodes()
    } catch (err: any) {
      console.error('Failed to reserve node:', err)
      throw err
    } finally {
      reserveNodeLoading.value = false
    }
  }

  // Unreserve a node
  async function unreserveNode(contractId: string) {
    await userService.unreserveNode(contractId)
    // Optimistically remove the node from the list
    rentedNodes.value = rentedNodes.value.filter(node => node.rentContractId?.toString() !== contractId)
  }

  // Add node to deployment
  async function addNodeToDeployment(deploymentName: string, nodePayload: { nodeId: number, role: string, vcpu: number, ram: number, storage: number }) {
    try {
      const response = await userService.addNodeToDeployment(deploymentName, nodePayload)
      // Optionally refresh data or handle response
      return response
    } catch (err: any) {
      console.error('Failed to add node to deployment:', err)
      throw err
    }
  }

  // Remove node from deployment
  async function removeNodeFromDeployment(deploymentName: string, nodeName: string) {
    try {
      const response = await userService.removeNodeFromDeployment(deploymentName, nodeName)
      // Optionally refresh data or handle response
      return response
    } catch (err: any) {
      console.error('Failed to remove node from deployment:', err)
      throw err
    }
  }

  // Calculate total monthly cost of rented nodes
  const totalMonthlyCost = computed(() => {
    return rentedNodes.value
      .filter(node => typeof node.price_usd === 'number')
      .reduce((sum, node) => sum + (node.price_usd || 0), 0)
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
    addNodeToDeployment,
    removeNodeFromDeployment,
    totalMonthlyCost,
    healthyNodes,
    unhealthyNodes,
    reserveNodeLoading
  }
}
