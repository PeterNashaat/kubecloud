import { WorkflowStatus } from '@/types/ewf'
import { api, createWorkflowStatusChecker } from './api'
import type { ApiResponse } from './authService'
import type { ChargeBalanceResponse } from './stripeService'
import { useNotificationStore } from '@/stores/notifications'

export interface ReserveNodeRequest {
  // Add any required fields if needed
}

export interface ReserveNodeResponse {
  workflow_id: string
  node_id: number
  email: string
}
export interface UnreserveNodeResponse {
  workflow_id: string
  contract_id: number,
  email: string,
}
export interface ChargeBalanceRequest {
  card_type: string
  payment_method_id: string
  amount: number
}

export interface RedeemVoucherResponse {
  amount: number,
  email: string,
  voucher_code: string,
  workflow_id: string
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

export interface UserInvoice {
  id: number
  user_id: number
  total: number
  nodes: any[]
  tax: number
  created_at: string
}

export interface SshKey {
  ID: number
  name: string
  public_key: string
  created_at: string
  updated_at: string
}

export interface AddSshKeyRequest {
  name: string
  public_key: string
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
    const response = await api.post<ApiResponse<ReserveNodeResponse>>(`/v1/user/nodes/${nodeId}`, data, { requiresAuth: true, showNotifications: true })
    const workflowChecker = createWorkflowStatusChecker(response.data.data.workflow_id, { delay: 3000, interval: 1000 })
    const status = await workflowChecker.status
    if(status === WorkflowStatus.StatusCompleted){
      useNotificationStore().addNotification({
        title: 'Node Reserved',
        message: 'Node has been successfully reserved.',
        type: 'success',
        duration: 5000
      })
    }
    if (status === WorkflowStatus.StatusFailed) {
      useNotificationStore().addNotification({
        title: 'Node Reservation Failed',
        message: 'Failed to reserve node',
        type: 'error',
        duration: 5000
      })
      throw new Error('Failed to reserve node')
    }
    return response.data.data
  }

  // List reserved nodes
  async listReservedNodes() {
    return api.get('/v1/user/nodes', { requiresAuth: true })
  }

  // Unreserve a node
  async unreserveNode(contractId: string) {
    const response = await api.delete<ApiResponse<UnreserveNodeResponse>>(`/v1/user/nodes/unreserve/${contractId}`, { requiresAuth: true, showNotifications: true })
    const workflowChecker = createWorkflowStatusChecker(response.data.data.workflow_id, { delay: 3000, interval: 1000 })
    const status = await workflowChecker.status
    if (status === WorkflowStatus.StatusFailed) {
      useNotificationStore().addNotification({
        title: 'Node Unreservation Failed',
        message: 'Failed to unreserve node',
        type: 'error',
        duration: 5000
      })
      throw new Error('Failed to unreserve node')
    }
    if (status === WorkflowStatus.StatusCompleted) {
      useNotificationStore().addNotification({
        title: 'Node Unreservation Success',
        message: 'Node has been successfully unreserved.',
        type: 'success',
        duration: 5000
      })
    }
    return response.data.data
  }

  // Charge balance
  async chargeBalance(data: ChargeBalanceRequest) {
    const response = await api.post<ApiResponse<ChargeBalanceResponse>>('/v1/user/balance/charge', data, { requiresAuth: true, showNotifications: true })
    const workflowChecker = createWorkflowStatusChecker(response.data.data.workflow_id, { interval: 6000 })
    const status = await workflowChecker.status
    if (status === WorkflowStatus.StatusFailed) {
      useNotificationStore().addNotification({
        title: 'Charge Failed',
        message: 'Failed to charge balance',
        type: 'error',
        duration: 5000
      })
    }
    if (status === WorkflowStatus.StatusCompleted) {
      useNotificationStore().addNotification({
        title: 'Charge Success',
        message: 'Charge balance successful',
        type: 'success',
        duration: 5000
      })
    }
  }

  // Create a PaymentIntent (to be implemented)
  // Replace this stub with a real API call to your backend that returns { clientSecret: string }
  async createPaymentIntent({ amount }: { amount: number }): Promise<{ clientSecret: string }> {
    throw new Error('UserService.createPaymentIntent is not implemented. Please implement this method to call your backend and return { clientSecret: string }.')
  }

  // List all invoices for the current user
  async listUserInvoices(): Promise<UserInvoice[]> {
    const response = await api.get<{ data: { invoices: UserInvoice[] } }>(
      '/v1/user/invoice/',
      { requiresAuth: true, showNotifications: true, errorMessage: 'Failed to load invoices' }
    )
    return response.data.data.invoices
  }

  // Download a specific invoice as PDF
  async downloadInvoice(invoiceId: number): Promise<Blob> {
    // The backend should return application/pdf for this endpoint
    const response = await api.get(
      `/v1/user/invoice/${invoiceId}`,
      { requiresAuth: true, errorMessage: 'Failed to download invoice' }
    )
    return response.data as Blob
  }

  // Redeem a voucher
  async redeemVoucher(voucherCode: string) {
    const res = await api.put<ApiResponse<RedeemVoucherResponse>>(`/v1/user/redeem/${voucherCode}`, {}, {
      requiresAuth: true,
      showNotifications: true,
      errorMessage: 'Failed to redeem voucher'
    })
    const workflowChecker = createWorkflowStatusChecker(res.data.data.workflow_id, { interval: 6000 })
    const status = await workflowChecker.status
    if (status === WorkflowStatus.StatusFailed) {
      useNotificationStore().addNotification({
        title: 'Voucher Redemption Failed',
        message: 'Failed to redeem voucher',
        type: 'error',
        duration: 5000
      })
    }
  }

  // Fetch the user's current balance
  async fetchBalance(): Promise<number> {
    try {
    const response = await api.get<{ data: { balance_usd: number, debt_usd: number } }>(
      '/v1/user/balance',
      { requiresAuth: true, showNotifications: false }
    )
    const { balance_usd, debt_usd } = response.data.data
    return (balance_usd || 0) - (debt_usd || 0)
  }catch(e){
    useNotificationStore().addNotification({
      title: 'Error',
      message: 'Failed to fetch balance',
      type: 'error',
      duration: 5000
    })
    return 0
  }
  }

  // List all SSH keys for the current user
  async listSshKeys(): Promise<SshKey[]> {
    const response = await api.get<ApiResponse<SshKey[]>>('/v1/user/ssh-keys', {
      requiresAuth: true,
      showNotifications: false
    })
    return response.data.data
  }

  // Add a new SSH key
  async addSshKey(data: AddSshKeyRequest): Promise<SshKey> {
    const response = await api.post<ApiResponse<SshKey>>('/v1/user/ssh-keys', data, {
      requiresAuth: true,
      showNotifications: true,
      successMessage: 'SSH key added successfully',
      errorMessage: 'Failed to add SSH key'
    })
    return response.data.data
  }

  // Delete an SSH key by ID
  async deleteSshKey(id: number): Promise<void> {
    await api.delete(`/v1/user/ssh-keys/${id}`, {
      requiresAuth: true,
      showNotifications: true,
      successMessage: 'SSH key deleted successfully',
      errorMessage: 'Failed to delete SSH key'
    })
  }
  // Add node to deployment
  async addNodeToDeployment(deploymentName: string, nodePayload: { nodeId: number, role: string, vcpu: number, ram: number, storage: number }) {
    return api.post(`/v1/deployments/${deploymentName}/nodes`, nodePayload, { requiresAuth: true, showNotifications: true })
  }

  // Remove node from deployment
  async removeNodeFromDeployment(deploymentName: string, nodeName: string) {
    return api.delete(`/v1/deployments/${deploymentName}/nodes/${nodeName}`, { requiresAuth: true, showNotifications: true })
  }
}

export const userService = new UserService()
