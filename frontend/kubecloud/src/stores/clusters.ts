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

export interface CreateClusterRequest {
  name: string
  region: string
  nodes: number
  nodeType: string
  tags?: string[]
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

  const createCluster = async (clusterData: CreateClusterRequest) => {
    // Implementation of createCluster function
    // This is a placeholder and should be replaced with the actual implementation
    // For example, you might use a third-party service or a custom implementation
    // to create a cluster and return a task ID and status
    return { task_id: 'someTaskId', status: 'deploying' }
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

  const updateCluster = async (clusterName: string, updates: Partial<Cluster>) => {
    const cluster = clusters.value.find(c => c.project_name === clusterName)
    if (!cluster) throw new Error('Cluster not found')

    try {
      // Real API call
      await api.put(`/clusters/${clusterName}`, updates)
      Object.assign(cluster, updates, { updated_at: new Date().toISOString() })
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update cluster'
      throw err
    }
  }

  const selectCluster = (cluster: Cluster | null) => {
    selectedCluster.value = cluster
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

  const initializeClusters = () => {
    // No mock data initialization
  }

  const deployCluster = async (payload: any) => {
    // Implementation of deployCluster function
    // This is a placeholder and should be replaced with the actual implementation
    // For example, you might use a third-party service or a custom implementation
    // to deploy a cluster and return a task ID and status
    return { task_id: 'someTaskId', status: 'deploying' }
  }

  const listenToEvents = async (taskId: string, callback: (data: any) => void) => {
    // Implementation of listenToEvents function
    // This is a placeholder and should be replaced with the actual implementation
    // For example, you might use a third-party service or a custom implementation
    // to listen to deployment events and call the callback with event data
  }

  const deploy = async (payload: any) => {
    const res = await deployCluster(payload)
    deploymentTaskId.value = res.task_id
    deploymentStatus.value = res.status
    deploymentEvents.value = []
    await listenToEvents(res.task_id, (data) => {
      deploymentEvents.value.push(data)
      // Optionally update deploymentStatus based on event data
    })
  }

  const listenToDeploymentEvents = async (taskId: string) => {
    await listenToEvents(taskId, (data) => {
      deploymentEvents.value.push(data)
      // Optionally update deploymentStatus based on event data
    })
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
    createCluster,
    deleteCluster,
    updateCluster,
    selectCluster,
    getClusterMetrics,
    initializeClusters,
    deploy,
    listenToDeploymentEvents,
    getClusterByName,
  }
}) 