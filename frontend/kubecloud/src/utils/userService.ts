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

  // Create a PaymentIntent (to be implemented)
  // Replace this stub with a real API call to your backend that returns { clientSecret: string }
  async createPaymentIntent({ amount }: { amount: number }): Promise<{ clientSecret: string }> {
    throw new Error('UserService.createPaymentIntent is not implemented. Please implement this method to call your backend and return { clientSecret: string }.')
  }
}

export const userService = new UserService() 