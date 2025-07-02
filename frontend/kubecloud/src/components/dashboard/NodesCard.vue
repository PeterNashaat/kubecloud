<template>
  <div class="nodes-card">
    <!-- Header Section -->
    <div class="card-header">
      <div class="header-content">
        <h2 class="card-title kubecloud-gradient kubecloud-glow-blue">
          My Nodes
        </h2>
        <p class="card-description">
          Manage your rented nodes and their resources.
        </p>
      </div>
      <div class="header-actions">
        <v-btn
          color="primary"
          variant="elevated"
          prepend-icon="mdi-plus"
          @click="navigateToReserve"
        >
          Reserve New Node
        </v-btn>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="stats-section">
      <v-row>
        <v-col cols="12" sm="6" md="3">
          <v-card class="stat-card" flat>
            <div class="stat-content">
              <div class="stat-icon">
                <v-icon size="32" color="primary">mdi-server-network</v-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ total }}</div>
                <div class="stat-label">Total Nodes</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-card class="stat-card" flat>
            <div class="stat-content">
              <div class="stat-icon">
                <v-icon size="32" color="success">mdi-check-circle</v-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ healthyNodes.length }}</div>
                <div class="stat-label">Healthy</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-card class="stat-card" flat>
            <div class="stat-content">
              <div class="stat-icon">
                <v-icon size="32" color="warning">mdi-alert-circle</v-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ unhealthyNodes.length }}</div>
                <div class="stat-label">Unhealthy</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-card class="stat-card" flat>
            <div class="stat-content">
              <div class="stat-icon">
                <v-icon size="32" color="info">mdi-currency-usd</v-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">${{ totalMonthlyCost.toFixed(2) }}</div>
                <div class="stat-label">Monthly Cost</div>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-section">
      <v-progress-circular
        indeterminate
        color="primary"
        size="64"
      />
      <p class="loading-text">Loading your nodes...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-section">
      <v-icon size="64" color="error" class="mb-4">mdi-alert-circle</v-icon>
      <h3>Failed to load nodes</h3>
      <p>{{ error }}</p>
      <v-btn
        color="primary"
        variant="outlined"
        @click="fetchRentedNodes"
      >
        Try Again
      </v-btn>
    </div>

    <!-- Empty State -->
    <div v-else-if="rentedNodes.length === 0" class="empty-section">
      <v-icon size="64" color="primary" class="mb-4">mdi-server-network</v-icon>
      <h3>No nodes rented yet</h3>
      <p>Start by reserving your first node to deploy your applications.</p>
      <v-btn
        color="primary"
        variant="elevated"
        @click="navigateToReserve"
      >
        Reserve Your First Node
      </v-btn>
    </div>

    <!-- Nodes Grid -->
    <div v-else class="nodes-section">
      <div class="nodes-grid">
        <div v-for="node in rentedNodes" :key="node.id" class="node-card">
          <div class="node-header">
            <div class="node-title-section">
              <h3 class="node-title">{{ node.name || `Node #${node.id}` }}</h3>
              <div class="node-status">
                <v-chip
                  :color="node.healthy ? 'success' : 'error'"
                  size="small"
                  variant="elevated"
                >
                  <v-icon size="16" class="mr-1">
                    {{ node.healthy ? 'mdi-check-circle' : 'mdi-alert-circle' }}
                  </v-icon>
                  {{ node.healthy ? 'Healthy' : 'Unhealthy' }}
                </v-chip>
              </div>
            </div>
            <div class="node-price">${{ node.price?.toFixed(2) ?? 'N/A' }}/month</div>
          </div>

          <div class="node-location" v-if="node.location">
            <v-icon size="16" class="mr-1">mdi-map-marker</v-icon>
            {{ node.location }}
          </div>

          <div class="node-specs">
            <div class="spec-item">
              <span class="spec-label">CPU:</span>
              <v-progress-linear
                :model-value="getAvailableCPU(node)"
                :max="getTotalCPU(node)"
                height="16"
                color="primary"
                rounded
                class="mb-1 mr-2"
                style="width: 120px; display: inline-block; vertical-align: middle;"
              >
                <template #default>
                  <span style="font-size: 0.95em;">{{ getAvailableCPU(node) }} / {{ getTotalCPU(node) }} vCPU</span>
                </template>
              </v-progress-linear>
            </div>
            <div class="spec-item">
              <span class="spec-label">RAM:</span>
              <v-progress-linear
                :model-value="getAvailableRAM(node)"
                :max="getTotalRAM(node)"
                height="16"
                color="success"
                rounded
                class="mb-1 mr-2"
                style="width: 120px; display: inline-block; vertical-align: middle;"
              >
                <template #default>
                  <span style="font-size: 0.95em;">{{ getAvailableRAM(node) }} / {{ getTotalRAM(node) }} GB</span>
                </template>
              </v-progress-linear>
            </div>
            <div class="spec-item">
              <span class="spec-label">Storage:</span>
              <v-progress-linear
                :model-value="getAvailableStorage(node)"
                :max="getTotalStorage(node)"
                height="16"
                color="info"
                rounded
                class="mb-1 mr-2"
                style="width: 120px; display: inline-block; vertical-align: middle;"
              >
                <template #default>
                  <span style="font-size: 0.95em;">{{ getAvailableStorage(node) }} / {{ getTotalStorage(node) }} GB</span>
                </template>
              </v-progress-linear>
            </div>
          </div>

          <div class="node-chips">
            <v-chip v-if="node.gpu || (node.gpus && node.gpus.length > 0)" color="deep-purple-accent-2" text-color="white" size="small" variant="elevated">
              <v-icon size="16" class="mr-1">mdi-nvidia</v-icon>
              GPU
            </v-chip>
            <v-chip v-if="node.dedicated" color="orange" text-color="white" size="small" variant="elevated">
              <v-icon size="16" class="mr-1">mdi-star</v-icon>
              Dedicated
            </v-chip>
            <v-chip v-if="node.rentContractId" color="blue" text-color="white" size="small" variant="elevated">
              <v-icon size="16" class="mr-1">mdi-file-document</v-icon>
              Contract #{{ node.rentContractId }}
            </v-chip>
          </div>

          <div class="node-actions">
            <v-btn
              color="error"
              variant="outlined"
              size="small"
              @click="confirmUnreserve(node)"
              :loading="unreservingNode === node.rentContractId?.toString()"
            >
              <v-icon size="16" class="mr-1">mdi-close</v-icon>
              Unreserve
            </v-btn>
          </div>
        </div>
      </div>
    </div>

    <!-- Unreserve Confirmation Dialog -->
    <v-dialog v-model="showUnreserveDialog" max-width="400">
      <v-card>
        <v-card-title>Confirm Unreservation</v-card-title>
        <v-card-text>
          Are you sure you want to unreserve this node? This action cannot be undone and will cancel your rental contract.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            color="grey"
            variant="text"
            @click="showUnreserveDialog = false"
          >
            Cancel
          </v-btn>
          <v-btn
            color="error"
            variant="elevated"
            @click="handleUnreserve"
            :loading="unreservingNode === selectedNode?.rentContractId?.toString()"
          >
            Unreserve
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNodeManagement, type RentedNode } from '../../composables/useNodeManagement'

const router = useRouter()
const {
  rentedNodes,
  total,
  loading,
  error,
  fetchRentedNodes,
  unreserveNode,
  totalMonthlyCost,
  healthyNodes,
  unhealthyNodes
} = useNodeManagement()

// Dialog state
const showUnreserveDialog = ref(false)
const selectedNode = ref<RentedNode | null>(null)
const unreservingNode = ref<string | null>(null)

onMounted(() => {
  fetchRentedNodes()
})

const navigateToReserve = () => {
  router.push('/nodes')
}

const confirmUnreserve = (node: RentedNode) => {
  selectedNode.value = node
  showUnreserveDialog.value = true
}

const handleUnreserve = async () => {
  if (!selectedNode.value?.rentContractId) return
  
  unreservingNode.value = selectedNode.value.rentContractId.toString()
  try {
    await unreserveNode(selectedNode.value.rentContractId.toString())
    showUnreserveDialog.value = false
    selectedNode.value = null
  } catch (err) {
    console.error('Failed to unreserve node:', err)
  } finally {
    unreservingNode.value = null
  }
}

// Resource calculation functions
function getTotalCPU(node: RentedNode) {
  return node.total_resources?.cru ?? node.resources?.cpu ?? 0
}

function getUsedCPU(node: RentedNode) {
  return node.used_resources?.cru ?? 0
}

function getAvailableCPU(node: RentedNode) {
  return Math.max(getTotalCPU(node) - getUsedCPU(node), 0)
}

function getTotalRAM(node: RentedNode) {
  return node.total_resources?.mru ? Math.round(node.total_resources.mru / (1024 * 1024 * 1024)) : (node.resources?.memory ?? 0)
}

function getUsedRAM(node: RentedNode) {
  return node.used_resources?.mru ? Math.round(node.used_resources.mru / (1024 * 1024 * 1024)) : 0
}

function getAvailableRAM(node: RentedNode) {
  return Math.max(getTotalRAM(node) - getUsedRAM(node), 0)
}

function getTotalStorage(node: RentedNode) {
  return node.total_resources?.sru ? Math.round(node.total_resources.sru / (1024 * 1024 * 1024)) : (node.resources?.storage ?? 0)
}

function getUsedStorage(node: RentedNode) {
  return node.used_resources?.sru ? Math.round(node.used_resources.sru / (1024 * 1024 * 1024)) : 0
}

function getAvailableStorage(node: RentedNode) {
  return Math.max(getTotalStorage(node) - getUsedStorage(node), 0)
}
</script>

<style scoped>
.nodes-card {
  background: rgba(10, 25, 47, 0.85);
  border: 1px solid rgba(96, 165, 250, 0.15);
  border-radius: 1rem;
  padding: 2rem;
  color: #CBD5E1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  gap: 1rem;
}

.header-content {
  flex: 1;
}

.card-title {
  font-size: 1.75rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  line-height: 1.2;
}

.card-description {
  color: #94A3B8;
  font-size: 1rem;
  line-height: 1.5;
  margin: 0;
}

.stats-section {
  margin-bottom: 2rem;
}

.stat-card {
  background: rgba(15, 23, 42, 0.6) !important;
  border: 1px solid rgba(96, 165, 250, 0.1) !important;
  border-radius: 0.75rem !important;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.stat-card:hover {
  border-color: rgba(96, 165, 250, 0.3) !important;
  transform: translateY(-2px);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon {
  background: rgba(96, 165, 250, 0.1);
  border: 1px solid rgba(96, 165, 250, 0.2);
  border-radius: 0.5rem;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: #F8FAFC;
  line-height: 1;
}

.stat-label {
  font-size: 0.875rem;
  color: #94A3B8;
  margin-top: 0.25rem;
}

.loading-section,
.error-section,
.empty-section {
  text-align: center;
  padding: 4rem 2rem;
}

.loading-text {
  margin-top: 1rem;
  color: #94A3B8;
}

.nodes-section {
  margin-top: 2rem;
}

.nodes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
}

.node-card {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(96, 165, 250, 0.1);
  border-radius: 0.75rem;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.node-card:hover {
  border-color: rgba(96, 165, 250, 0.3);
  transform: translateY(-2px);
}

.node-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.node-title-section {
  flex: 1;
}

.node-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #F8FAFC;
  margin: 0 0 0.5rem 0;
  line-height: 1.2;
}

.node-status {
  margin-bottom: 0.5rem;
}

.node-price {
  font-size: 1.125rem;
  font-weight: 600;
  color: #10B981;
  text-align: right;
}

.node-location {
  display: flex;
  align-items: center;
  color: #94A3B8;
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.node-specs {
  margin-bottom: 1rem;
}

.spec-item {
  display: flex;
  align-items: center;
  margin-bottom: 0.75rem;
}

.spec-label {
  font-size: 0.875rem;
  color: #94A3B8;
  min-width: 60px;
  margin-right: 0.5rem;
}

.node-chips {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.node-actions {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: stretch;
  }

  .header-actions {
    align-self: stretch;
  }

  .nodes-grid {
    grid-template-columns: 1fr;
  }

  .node-header {
    flex-direction: column;
    align-items: stretch;
  }

  .node-price {
    text-align: left;
    margin-top: 0.5rem;
  }
}
</style> 