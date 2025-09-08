import { ref } from 'vue'
import { userService } from '../utils/userService'
import type { RawNode } from '../types/rawNode'

export interface GridNodeFilters {
  healthy?: boolean
  size?: number
  page?: number
}

export function useGridNodes() {
  const gridNodes = ref<RawNode[]>([])
  const total = ref<number>(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Fetch all grid nodes from the public endpoint
  async function fetchGridNodes(filters?: GridNodeFilters) {
    loading.value = true
    error.value = null

    try {
      const response = await userService.listAllGridNodes(filters)
      const responseData = response.data as any

      if (responseData.data?.nodes) {
        gridNodes.value = responseData.data.nodes
        total.value = responseData.data.total || 0
      } else {
        gridNodes.value = []
        total.value = 0
      }
    } catch (err: any) {
      console.error('Failed to fetch grid nodes from /v1/nodes endpoint:', err)
      error.value = err?.message || 'Failed to fetch grid nodes'
      gridNodes.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  return {
    gridNodes,
    total,
    loading,
    error,
    fetchGridNodes
  }
}
