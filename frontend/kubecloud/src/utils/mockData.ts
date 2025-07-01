// Mock data for clusters
import type { Cluster } from '../stores/clusters'

export const mockClusters: Cluster[] = [
  {
    id: '1',
    name: 'Production Cluster',
    status: 'running',
    region: 'US East',
    nodes: 3,
    cpu: '12',
    memory: '24 GB',
    storage: '1 TB',
    createdAt: '2023-07-01',
    lastUpdated: '2023-07-10',
    cost: 125,
    tags: ['production', 'critical']
  },
  {
    id: '2',
    name: 'Staging Cluster',
    status: 'stopped',
    region: 'US West',
    nodes: 2,
    cpu: '8',
    memory: '16 GB',
    storage: '500 GB',
    createdAt: '2023-06-15',
    lastUpdated: '2023-07-05',
    cost: 80,
    tags: ['staging']
  }
]

// Simple mockApi implementation
export const mockApi = {
  get: async (url: string) => {
    if (url === '/clusters') {
      return { data: mockClusters }
    }
    throw new Error('Not implemented in mockApi')
  },
  post: async (url: string, data?: any) => {
    return { data: {} }
  },
  put: async (url: string, data?: any) => {
    return { data: {} }
  },
  delete: async (url: string) => {
    return { data: {} }
  }
}

export const MOCK_CONFIG = {
  enabled: true
} 