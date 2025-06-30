import { ref } from 'vue'
import { api } from '../utils/api'

export interface Node {
  id: number
  name?: string
  location?: string
  resources?: {
    cpu: number
    memory: number
    storage: number
  }
  gpu?: string
  price?: number
  // Add other fields as needed from backend
}

export function useNodes() {
  const nodes = ref<Node[]>([])
  const total = ref<number>(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchNodes() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get<{ total: number; nodes: Node[] }>('/v1/user/nodes', { requiresAuth: true })
      nodes.value = response.data.nodes
      total.value = response.data.total
    } catch (err: any) {
      error.value = err?.message || 'Failed to fetch nodes'
    } finally {
      loading.value = false
    }
  }

  return {
    nodes,
    total,
    loading,
    error,
    fetchNodes
  }
} 