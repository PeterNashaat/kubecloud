import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../utils/api'

export interface Cluster {
  id: number
  project_name: string
  cluster: {
    name: string
    status?: string
    region?: string
    nodes?: number
    cpu?: string
    memory?: string
    storage?: string
    tags?: string[]
    [key: string]: any
  }
  created_at: string
  updated_at: string
}

export interface ClusterMetrics {
  cpuUsage: number
  memoryUsage: number
  storageUsage: number
  networkIn: number
  networkOut: number
  activeConnections: number
}

export const useClusterStore = defineStore('clusters', () => {
  // State
  const clusters = ref<Cluster[]>([])
  const selectedCluster = ref<Cluster | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const deploymentTaskId = ref<string | null>(null)
  const deploymentStatus = ref<string | null>(null)
  const deploymentEvents = ref<any[]>([])

  // Computed properties
  const clustersByRegion = computed(() => {
    const grouped: Record<string, Cluster[]> = {}
    clusters.value.forEach(cluster => {
      const region = cluster.cluster.region
      if (!region) return // skip clusters with undefined region
      if (!grouped[region]) {
        grouped[region] = []
      }
      grouped[region].push(cluster)
    })
    return grouped
  })

  // Actions
  const fetchClusters = async () => {
    isLoading.value = true
    error.value = null

    try {
      const response = await api.get('/v1/deployments', { requiresAuth: true })
      clusters.value = (response.data as { deployments: Cluster[] }).deployments
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch clusters'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const deleteCluster = async (name: string) => {
    isLoading.value = true
    error.value = null

    try {
      await api.delete(`/v1/deployments/${name}`, { requiresAuth: true })
      clusters.value = clusters.value.filter(cluster => cluster.project_name !== name)
      
      if (selectedCluster.value?.project_name === name) {
        selectedCluster.value = null
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete cluster'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const getClusterMetrics = async (clusterId: string): Promise<ClusterMetrics> => {
    try {
      // Real API call (replace with actual endpoint if available)
      const response = await api.get(`/clusters/${clusterId}/metrics`)
      return response.data as ClusterMetrics
    } catch (err) {
      throw new Error('Failed to fetch cluster metrics')
    }
  }

  const getClusterByName = async (name: string): Promise<Cluster | null> => {
    isLoading.value = true
    error.value = null
    try {
      const response = await api.get(`/v1/deployments/${name}`, { requiresAuth: true })
      return response.data as Cluster
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch cluster'
      return null
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Add nodes to a cluster (replaces the cluster's node list)
   * @param clusterName The name of the cluster
   * @param clusterObject The full cluster object with the updated nodes array
   * @returns Promise<ApiResponse<any>>
   */
  const addNodesToCluster = async (clusterName: string, clusterObject: any) => {
    return api.post(`/v1/deployments/${clusterName}/nodes`, clusterObject, { requiresAuth: true })
  }

  /**
   * Remove a node from a cluster
   * @param clusterName The name of the cluster
   * @param nodeName The name of the node to remove
   * @returns Promise<ApiResponse<any>>
   */
  const removeNodeFromCluster = async (clusterName: string, nodeName: string) => {
    return api.delete(`/v1/deployments/${clusterName}/nodes/${nodeName}`, { requiresAuth: true })
  }

  return {
    // State
    clusters: computed(() => clusters.value),
    selectedCluster: computed(() => selectedCluster.value),
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),
    deploymentTaskId: computed(() => deploymentTaskId.value),
    deploymentStatus: computed(() => deploymentStatus.value),
    deploymentEvents: computed(() => deploymentEvents.value),

    // Computed
    clustersByRegion,

    // Actions
    fetchClusters,
    deleteCluster,
    getClusterMetrics,
    addNodesToCluster,
    removeNodeFromCluster,
  }
}) 