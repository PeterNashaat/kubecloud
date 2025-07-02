import { api } from './api'

export interface ReserveNodeRequest {
  // Add any required fields if needed
}

export interface ChargeBalanceRequest {
  card_type: string
  payment_method_id: string
  amount: number
}

export interface Node {
  id: number
  node_id: number
  farm_id: number
  twin_id: number
  country: string
  city: string
  latitude: number
  longitude: number
  created: number
  farming_policy_id: number
  interfaces: any[]
  secure: boolean
  virtualized: boolean
  serial_number: string
  created_at: number
  updated_at: number
  location_id: string
  power: {
    state: string
    target: string
  }
  public_config: {
    ipv4: string
    ipv6: string
    gw4: string
    gw6: string
    domain: string
  }
  public_ips: any[]
  resources: {
    hru: number
    sru: number
    cru: number
    mru: number
  }
  location: {
    country: string
    city: string
    latitude: number
    longitude: number
  }
  status: string
  healthy: boolean
  rentable: boolean
  rented: boolean
  rented_by: number
  rent_contract_id: number
  certification_type: string
  dedicated: boolean
  grid_version: number
}

export interface NodesResponse {
  total: number
  nodes: Node[]
}

export class UserService {
  // List all available nodes
  async listNodes(filters?: any) {
    const queryParams = new URLSearchParams()
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          queryParams.append(key, String(value))
        }
      })
    }
    
    const endpoint = `/v1/nodes${queryParams.toString() ? `?${queryParams.toString()}` : ''}`
    return api.get<NodesResponse>(endpoint, { 
      requiresAuth: true,
      showNotifications: false // Don't show notifications for node listing
    })
  }

  // Reserve a node
  async reserveNode(nodeId: number, data: ReserveNodeRequest = {}) {
    return api.post(`/v1/user/nodes/${nodeId}`, data, { requiresAuth: true, showNotifications: true })
  }

  // List reserved nodes
  async listReservedNodes() {
    return api.get('/v1/user/nodes', { requiresAuth: true })
  }

  // Unreserve a node
  async unreserveNode(contractId: string) {
    return api.post(`/v1/user/nodes/unreserve/${contractId}`, {}, { requiresAuth: true, showNotifications: true })
  }

  // Charge balance
  async chargeBalance(data: ChargeBalanceRequest) {
    return api.post('/v1/user/charge_balance', data, { requiresAuth: true, showNotifications: true })
  }

  // Create a PaymentIntent (to be implemented)
  // Replace this stub with a real API call to your backend that returns { clientSecret: string }
  async createPaymentIntent({ amount }: { amount: number }): Promise<{ clientSecret: string }> {
    throw new Error('UserService.createPaymentIntent is not implemented. Please implement this method to call your backend and return { clientSecret: string }.')
  }
}

export const userService = new UserService() 