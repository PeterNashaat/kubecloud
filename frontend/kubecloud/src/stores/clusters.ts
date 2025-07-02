import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../utils/api'

export interface Cluster {
  id: string
  name: string
  status: 'running' | 'stopped' | 'starting' | 'stopping' | 'error'
  region: string
  nodes: number
  cpu: string
  memory: string
  storage: string
  createdAt: string
  lastUpdated: string
  cost: number
  tags: string[]
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
  const runningClusters = computed(() => 
    clusters.value.filter(cluster => cluster.status === 'running')
  )

  const stoppedClusters = computed(() => 
    clusters.value.filter(cluster => cluster.status === 'stopped')
  )

  const totalCost = computed(() => 
    clusters.value.reduce((sum, cluster) => sum + cluster.cost, 0)
  )

  const clustersByRegion = computed(() => {
    const grouped: Record<string, Cluster[]> = {}
    clusters.value.forEach(cluster => {
      if (!grouped[cluster.region]) {
        grouped[cluster.region] = []
      }
      grouped[cluster.region].push(cluster)
    })
    return grouped
  })

  // Actions
  const fetchClusters = async () => {
    isLoading.value = true
    error.value = null

    try {
      // Real API call
      const response = await api.get('/clusters')
      clusters.value = response.data as Cluster[]
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch clusters'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const createCluster = async (clusterData: CreateClusterRequest) => {
    isLoading.value = true
    error.value = null

    try {
      // Real API call
      const response = await api.post('/clusters', clusterData)
      const newCluster: Cluster = {
        ...(response.data as Cluster),
        ...clusterData,
        status: 'starting',
        createdAt: new Date().toISOString(),
        lastUpdated: new Date().toISOString(),
        cost: 0, // Will be calculated once running
      }
      clusters.value.push(newCluster)
      return newCluster
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create cluster'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const deleteCluster = async (clusterId: string) => {
    isLoading.value = true
    error.value = null

    try {
      // Real API call
      await api.delete(`/clusters/${clusterId}`)
      clusters.value = clusters.value.filter(cluster => cluster.id !== clusterId)
      
      if (selectedCluster.value?.id === clusterId) {
        selectedCluster.value = null
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete cluster'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const startCluster = async (clusterId: string) => {
    const cluster = clusters.value.find(c => c.id === clusterId)
    if (!cluster) throw new Error('Cluster not found')

    cluster.status = 'starting'
    
    try {
      // Real API call
      await api.post(`/clusters/${clusterId}/start`, { action: 'start' })
      // Optionally update status after API call
      cluster.status = 'running'
      cluster.lastUpdated = new Date().toISOString()
    } catch (err) {
      cluster.status = 'error'
      throw err
    }
  }

  const stopCluster = async (clusterId: string) => {
    const cluster = clusters.value.find(c => c.id === clusterId)
    if (!cluster) throw new Error('Cluster not found')

    cluster.status = 'stopping'
    
    try {
      // Real API call
      await api.post(`/clusters/${clusterId}/stop`, { action: 'stop' })
      // Optionally update status after API call
      cluster.status = 'stopped'
      cluster.lastUpdated = new Date().toISOString()
    } catch (err) {
      cluster.status = 'error'
      throw err
    }
  }

  const updateCluster = async (clusterId: string, updates: Partial<Cluster>) => {
    const cluster = clusters.value.find(c => c.id === clusterId)
    if (!cluster) throw new Error('Cluster not found')

    try {
      // Real API call
      await api.put(`/clusters/${clusterId}`, updates)
      Object.assign(cluster, updates, { lastUpdated: new Date().toISOString() })
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
    runningClusters,
    stoppedClusters,
    totalCost,
    clustersByRegion,

    // Actions
    fetchClusters,
    createCluster,
    deleteCluster,
    startCluster,
    stopCluster,
    updateCluster,
    selectCluster,
    getClusterMetrics,
    initializeClusters,
    deploy,
    listenToDeploymentEvents,
  }
}) 