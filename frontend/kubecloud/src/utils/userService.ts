import { api } from './api'

export interface ReserveNodeRequest {
  // Add any required fields if needed
}

export interface ChargeBalanceRequest {
  card_type: string
  payment_method_id: string
  amount: number
}

export class UserService {
  // Reserve a node
  async reserveNode(nodeId: number, data: ReserveNodeRequest = {}) {
    return api.post(`/v1/user/nodes/${nodeId}`, data, { requiresAuth: true, showNotifications: true })
  }

  // List reserved nodes
  async listReservedNodes() {
    return api.get('/v1/user/nodes/reserved', { requiresAuth: true })
  }

  // Unreserve a node
  async unreserveNode(contractId: string) {
    return api.post(`/v1/user/nodes/unreserve/${contractId}`, {}, { requiresAuth: true, showNotifications: true })
  }

  // Charge balance
  async chargeBalance(data: ChargeBalanceRequest) {
    return api.post('/v1/user/charge_balance', data, { requiresAuth: true, showNotifications: true })
  }
}

export const userService = new UserService() 