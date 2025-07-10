<template>
  <div class="manage-cluster-container">
    <div class="container">
      <v-container fluid class="pa-0">
        <div v-if="loading" class="d-flex justify-center align-center" style="min-height: 60vh;">
          <v-progress-circular indeterminate color="primary" size="48" />
        </div>
        <div v-else-if="notFound" class="d-flex flex-column justify-center align-center" style="min-height: 60vh;">
          <h2>Cluster Not Found</h2>
          <v-btn color="primary" @click="goBack">Back to Dashboard</v-btn>
        </div>
        <div v-else-if="cluster" class="manage-header mb-6">
          <div class="manage-header-content">
            <h1 class="manage-title">{{ cluster?.cluster?.name || '-' }}</h1>
            <p class="manage-subtitle">Manage your Kubernetes cluster configuration and resources</p>
            <div class="tag-list mt-2">
            </div>
          </div>
        </div>
        <div v-if="!loading && !notFound && cluster" class="manage-content-wrapper">
              <div class="status-actions align-end">
                <v-btn variant="outlined" class="btn btn-outline" @click="openKubeconfigModal">
                  <v-icon icon="mdi-eye" class="mr-2"></v-icon>
                  Show Kubeconfig
                </v-btn>
                <v-btn variant="outlined" class="btn btn-outline" @click="openEditClusterNodesDialog">
                  <v-icon icon="mdi-pencil" class="mr-2"></v-icon>
                  Edit Cluster
                </v-btn>
                <v-btn variant="outlined" class="btn btn-outline" color="error" @click="openDeleteModal">
                  <v-icon icon="mdi-delete" class="mr-2"></v-icon>
                  Delete
                </v-btn>
              </div>

          <!-- Tabs for Overview, Nodes, Metrics, Events -->
          <v-tabs v-model="tab" class="mb-4">
            <v-tab value="overview">Overview</v-tab>
            <v-tab value="nodes">Nodes</v-tab>
            <v-tab value="metrics">Metrics</v-tab>
            <v-tab value="events">Events</v-tab>
          </v-tabs>

          <div class="card main-content-card">
            <div class="tab-content">
              <!-- Overview Tab -->
              <div v-if="tab === 'overview'">
                <div class="overview-grid">
                  <div class="card overview-card">
                    <h3 class="dashboard-card-title">
                      <v-icon icon="mdi-server" class="mr-2"></v-icon>
                      Cluster Resources
                    </h3>
                    <div class="resource-list">
                      <div class="resource-item">
                        <span class="resource-label">Nodes:</span>
                        <span class="resource-value">{{ filteredNodes.length }}</span>
                      </div>
                      <div class="resource-item">
                        <span class="resource-label">vCPU:</span>
                        <span class="resource-value">{{ totalVcpu }}</span>
                      </div>
                      <div class="resource-item">
                        <span class="resource-label">RAM:</span>
                        <span class="resource-value">{{ totalRam }} MB</span>
                      </div>
                      <div class="resource-item">
                        <span class="resource-label">Storage:</span>
                        <span class="resource-value">{{ totalStorage }} MB</span>
                      </div>
                    </div>
                  </div>
                  <div class="card overview-card details-card">
                    <h3 class="dashboard-card-title">
                      <v-icon icon="mdi-information" class="mr-2"></v-icon>
                      Cluster Details
                    </h3>
                    <div class="details-grid">
                      <div class="detail-item">
                        <span class="detail-label">Project Name:</span>
                        <span class="detail-value">{{ cluster.project_name || '-' }}</span>
                      </div>
                      <div class="detail-item">
                        <span class="detail-label">Cluster Name:</span>
                        <span class="detail-value">{{ cluster.cluster?.name || '-' }}</span>
                      </div>
                      <div class="detail-item">
                        <span class="detail-label">Created:</span>
                        <span class="detail-value">{{ formatDate(cluster.created_at) }}</span>
                      </div>
                      <div class="detail-item">
                        <span class="detail-label">Last Updated:</span>
                        <span class="detail-value">{{ formatDate(cluster.updated_at) }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Nodes Tab -->
              <div v-else-if="tab === 'nodes'">
                <h3 class="dashboard-card-title mb-4">
                  <v-icon icon="mdi-lan" class="mr-2"></v-icon>
                  Cluster Nodes
                </h3>
                <v-table v-if="filteredNodes.length">
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Type</th>
                      <th>CPU</th>
                      <th>RAM</th>
                      <th>Storage</th>
                      <th>IP</th>
                      <th>Mycelium IP</th>
                      <th>Planetary IP</th>
                      <th>Contract ID</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="node in filteredNodes" :key="node.node_id">
                      <td>{{ node.name }}</td>
                      <td>{{ node.type }}</td>
                      <td>{{ node.cpu }}</td>
                      <td>{{ node.memory }} MB</td>
                      <td>{{ node.root_size + node.disk_size }} MB</td>
                      <td>
                        <span class="truncate-cell">
                          {{ node.ip || '-' }}
                        </span>
                      </td>
                      <td>
                        <v-tooltip activator="parent" location="top" v-if="node.mycelium_ip">
                          <template #activator="{ props }">
                            <span class="truncate-cell" v-bind="props">
                              {{ node.mycelium_ip }}
                            </span>
                          </template>
                          <span>{{ node.mycelium_ip }}</span>
                        </v-tooltip>
                        <span v-else>-</span>
                      </td>
                      <td>
                        <v-tooltip activator="parent" location="top" v-if="node.planetary_ip">
                          <template #activator="{ props }">
                            <span class="truncate-cell" v-bind="props">
                              {{ node.planetary_ip }}
                            </span>
                          </template>
                          <span>{{ node.planetary_ip }}</span>
                        </v-tooltip>
                        <span v-else>-</span>
                      </td>
                      <td>{{ node.contract_id || '-' }}</td>
                    </tr>
                  </tbody>
                </v-table>
                <div v-else class="empty-message">No node details available.</div>
              </div>

              <!-- Metrics Tab -->
              <div v-else-if="tab === 'metrics'">
                <h3 class="dashboard-card-title mb-4">
                  <v-icon icon="mdi-chart-line" class="mr-2"></v-icon>
                  Live Metrics
                </h3>
                <v-alert v-if="metricsError" type="error" class="mb-4">{{ metricsError }}</v-alert>
                <v-progress-linear v-if="metricsLoading" indeterminate color="primary" class="mb-4" />
                <div v-if="metrics">
                  <div class="usage-metrics">
                    <div class="usage-item">
                      <div class="usage-header">
                        <span class="usage-label">CPU Usage</span>
                        <span class="usage-value">{{ metrics.cpuUsage }}%</span>
                      </div>
                      <v-progress-linear :model-value="metrics.cpuUsage" color="var(--color-primary)" height="8" rounded class="usage-bar" />
                    </div>
                    <div class="usage-item">
                      <div class="usage-header">
                        <span class="usage-label">Memory Usage</span>
                        <span class="usage-value">{{ metrics.memoryUsage }}%</span>
                      </div>
                      <v-progress-linear :model-value="metrics.memoryUsage" color="var(--color-primary)" height="8" rounded class="usage-bar" />
                    </div>
                    <div class="usage-item">
                      <div class="usage-header">
                        <span class="usage-label">Storage Usage</span>
                        <span class="usage-value">{{ metrics.storageUsage }}%</span>
                      </div>
                      <v-progress-linear :model-value="metrics.storageUsage" color="var(--color-primary)" height="8" rounded class="usage-bar" />
                    </div>
                    <div class="usage-item">
                      <div class="usage-header">
                        <span class="usage-label">Network In</span>
                        <span class="usage-value">{{ metrics.networkIn }} MB</span>
                      </div>
                    </div>
                    <div class="usage-item">
                      <div class="usage-header">
                        <span class="usage-label">Network Out</span>
                        <span class="usage-value">{{ metrics.networkOut }} MB</span>
                      </div>
                    </div>
                    <div class="usage-item">
                      <div class="usage-header">
                        <span class="usage-label">Active Connections</span>
                        <span class="usage-value">{{ metrics.activeConnections }}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-else-if="!metricsLoading && !metricsError" class="empty-message">No metrics available.</div>
              </div>

              <!-- Events Tab -->
              <div v-else-if="tab === 'events'">
                <h3 class="dashboard-card-title mb-4">
                  <v-icon icon="mdi-history" class="mr-2"></v-icon>
                  Deployment Events
                </h3>
                <v-alert v-if="eventsError" type="error" class="mb-4">{{ eventsError }}</v-alert>
                <v-progress-linear v-if="eventsLoading" indeterminate color="primary" class="mb-4" />
                <v-list v-if="events && events.length">
                  <v-list-item v-for="event in events" :key="event.id">
                    <v-list-item-content>
                      <v-list-item-title>{{ event.message }}</v-list-item-title>
                      <v-list-item-subtitle>{{ formatDate(event.timestamp) }}</v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-list>
                <div v-else-if="!eventsLoading && !eventsError" class="empty-message">No events found.</div>
              </div>
            </div>
          </div>
        </div>
      </v-container>
    </div>

    <!-- Kubeconfig Modal -->
    <KubeconfigDialog
      v-model="kubeconfigDialog"
      :projectName="projectName"
      :loading="kubeconfigLoading"
      :error="kubeconfigError"
      :content="kubeconfigContent"
      @copy="copyKubeconfig"
      @download="downloadKubeconfigFile"
    />

    <!-- Delete Confirmation Modal -->
    <DeleteClusterDialog
      v-model="showDeleteModal"
      :loading="deletingCluster"
      @confirm="confirmDelete"
    />

    <!-- Edit Cluster Nodes Modal -->
    <EditClusterNodesDialog
      v-model="editClusterNodesDialog"
      :cluster="cluster"
      :nodes="filteredNodes"
      :loading="nodesLoading"
      :available-nodes="availableNodes"
      :add-form-error="addFormError"
      :add-form-node="addFormNode"
      :can-assign-to-node="canAssignToNode"
      :add-node-loading="addNodeLoading"
      @add-node="addNode"
      @nodes-updated="editNodes = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useClusterStore } from '../../stores/clusters'
import { api } from '../../utils/api'
import { useNotificationStore } from '../../stores/notifications'
import { useNodeManagement, type RentedNode } from '../../composables/useNodeManagement'
import { getTotalCPU, getAvailableCPU, getTotalRAM, getAvailableRAM, getTotalStorage, getUsedStorage, getAvailableStorage } from '../../utils/nodeNormalizer';
import { formatDate } from '../../utils/dateUtils';
// Import dialogs
import EditClusterNodesDialog from './EditClusterNodesDialog.vue';
import KubeconfigDialog from './KubeconfigDialog.vue';
import DeleteClusterDialog from './DeleteClusterDialog.vue';

function getClusterUsedResources(nodeId: number) {
  // Sums up vcpu, ram, storage for all editNodes with this nodeId
  // editNodes may contain extended node objects with vcpu/ram/storage or cpu/memory/storage
  return (editNodes.value || []).filter((n: RentedNode) => n.nodeId === nodeId).reduce((acc: { vcpu: number, ram: number, storage: number }, n: RentedNode) => {
    acc.vcpu += ('vcpu' in n ? (n as any).vcpu : (n as any).cpu) || 0;
    acc.ram += ('ram' in n ? (n as any).ram : (n as any).memory) || 0;
    acc.storage += (n as any).storage || 0;
    return acc;
  }, { vcpu: 0, ram: 0, storage: 0 });
}

const router = useRouter()
const route = useRoute()
const clusterStore = useClusterStore()
const notificationStore = useNotificationStore()

const loading = ref(true)
const notFound = ref(false)

const projectName = computed(() => route.params.id?.toString() || '')
const cluster = computed(() =>
  clusterStore.clusters.find(c => c.project_name === projectName.value)
)

const tab = ref('overview')

const filteredNodes = computed(() => {
  if (Array.isArray(cluster.value?.cluster.nodes)) {
    return cluster.value.cluster.nodes.filter(node => typeof node === 'object' && node !== null)
  }
  return []
})

const totalVcpu = computed(() => {
  return filteredNodes.value.length
    ? filteredNodes.value.reduce((sum, node) => sum + (typeof node.cpu === 'number' ? node.cpu : 0), 0)
    : '-'
})
const totalRam = computed(() => {
  return filteredNodes.value.length
    ? filteredNodes.value.reduce((sum, node) => sum + (typeof node.memory === 'number' ? node.memory : 0), 0)
    : '-'
})
const totalStorage = computed(() => {
  return filteredNodes.value.length
    ? filteredNodes.value.reduce((sum, node) => sum + ((typeof node.root_size === 'number' ? node.root_size : 0) + (typeof node.disk_size === 'number' ? node.disk_size : 0)), 0)
    : '-'
})

const kubeconfigDialog = ref(false)
const kubeconfigContent = ref('')
const kubeconfigLoading = ref(false)
const kubeconfigError = ref('')

async function showKubeconfig() {
  kubeconfigLoading.value = true
  kubeconfigError.value = ''
  kubeconfigContent.value = ''
  try {
    const response = await api.get(`/v1/deployments/${projectName.value}/kubeconfig`, { requiresAuth: true, showNotifications: false })
    const data = response.data as { kubeconfig?: string }
    kubeconfigContent.value = data.kubeconfig || ''
  } catch (err: any) {
    kubeconfigError.value = err?.message || 'Failed to fetch kubeconfig'
  } finally {
    kubeconfigLoading.value = false
  }
}

function openKubeconfigModal() {
  kubeconfigDialog.value = true
  if (!kubeconfigContent.value && !kubeconfigLoading.value) {
    showKubeconfig()
  }
}

function copyKubeconfig() {
  if (kubeconfigContent.value) {
    navigator.clipboard.writeText(kubeconfigContent.value)
  }
}

function downloadKubeconfigFile() {
  if (!kubeconfigContent.value) return
  const blob = new Blob([kubeconfigContent.value], { type: 'application/x-yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${projectName.value}-kubeconfig.yaml`
  document.body.appendChild(a)
  a.click()
  setTimeout(() => {
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }, 100)
}

const showDeleteModal = ref(false)
const deletingCluster = ref(false)

async function confirmDelete() {
  deletingCluster.value = true
  showDeleteModal.value = false
  if (cluster.value) {
    await clusterStore.deleteCluster(cluster.value.project_name)
    router.push('/dashboard/clusters')
  }
  deletingCluster.value = false
}

function openDeleteModal() {
  showDeleteModal.value = true
}

const loadCluster = async () => {
  loading.value = true
  notFound.value = false
  try {
    if (!clusterStore.clusters.length) {
      await clusterStore.fetchClusters()
    }
    if (!cluster.value) {
      notFound.value = true
    }
  } catch (e) {
    notFound.value = true
  } finally {
    loading.value = false
  }
}

onMounted(loadCluster)
watch(() => projectName.value, loadCluster)

const goBack = () => {
  router.push('/dashboard')
}

// Actions
const metrics = ref<any>(null)
const metricsLoading = ref(false)
const metricsError = ref<string | null>(null)

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)
const eventsError = ref<string | null>(null)

// Metrics fetching
async function fetchMetrics() {
  metricsLoading.value = true
  metricsError.value = null
  try {
    metrics.value = await clusterStore.getClusterMetrics(projectName.value)
  } catch (e: any) {
    metricsError.value = e.message || 'Failed to fetch metrics'
  } finally {
    metricsLoading.value = false
  }
}

const editClusterNodesDialog = ref(false)

// Dummy state for masters/workers (replace with real cluster data)
const editNodes = ref<any[]>([])

const tableHeaders = [
  { title: 'Name', key: 'name' },
  { title: 'Type', key: 'type' },
  { title: 'CPU', key: 'cpu' },
  { title: 'RAM', key: 'memory' },
  { title: 'Storage', key: 'storage' },
  { title: 'IP', key: 'ip' },
  { title: 'Contract ID', key: 'contract_id' },
  { title: 'Actions', key: 'actions', sortable: false }
]

const editNodesWithStorage = computed(() =>
  (editNodes.value || []).map(node => ({
    ...node,
    storage: (node.root_size || 0) + (node.disk_size || 0)
  }))
)

const editTab = ref('list');

async function openEditClusterNodesDialog() {
  const nodesRaw = cluster.value?.cluster?.nodes;
  const nodes = Array.isArray(nodesRaw) ? nodesRaw : [];
  editNodes.value = nodes.map(n => ({ ...n }));
  editTab.value = 'list';
  editClusterNodesDialog.value = true;
  await fetchRentedNodes();
}

async function removeNode(nodeName: string) {
  if (!cluster.value || !cluster.value.cluster) return;
  await clusterStore.removeNodeFromCluster(cluster.value.cluster.name, nodeName)
  editNodes.value = editNodes.value.filter(n => n.name !== nodeName)
  await clusterStore.fetchClusters()
}
const addNodeLoading = ref(false)
const availableNodes = computed<RentedNode[]>(() => {
  return rentedNodes.value.filter((node: RentedNode) => {
    const clusterUsed = getClusterUsedResources(node.nodeId);
    const availCPU = getAvailableCPU(node) - clusterUsed.vcpu;
    const availRAM = getAvailableRAM(node) - clusterUsed.ram;
    const availStorage = getAvailableStorage(node) - clusterUsed.storage;
    return availCPU > 0 && availRAM > 0 && availStorage > 0;
  });
});

const { rentedNodes, loading: nodesLoading, fetchRentedNodes } = useNodeManagement()

const addFormNodeId = ref(null);
const addFormRole = ref('master');
const addFormVcpu = ref(1);
const addFormRam = ref(1);
const addFormStorage = ref(1);
const addFormError = ref('');

const addFormNode = computed<RentedNode | undefined>(() => availableNodes.value.find((n: RentedNode) => n.nodeId === addFormNodeId.value));

const canAssignToNode = computed(() => {
  const node = addFormNode.value;
  if (!node) return false;
  return (
    addFormVcpu.value > 0 &&
    addFormRam.value > 0 &&
    addFormStorage.value > 0 &&
    addFormVcpu.value <= getAvailableCPU(node) &&
    addFormRam.value <= getAvailableRAM(node) &&
    addFormStorage.value <= getAvailableStorage(node)
  );
});

watch([addFormNodeId, addFormVcpu, addFormRam, addFormStorage], () => {
  const node = addFormNode.value;
  if (!node) {
    addFormError.value = '';
    return;
  }
  if (
    addFormVcpu.value > getAvailableCPU(node) ||
    addFormRam.value > getAvailableRAM(node) ||
    addFormStorage.value > getAvailableStorage(node)
  ) {
    addFormError.value = 'Requested resources exceed available for the selected node.';
  } else {
    addFormError.value = '';
  }
});

async function addNode(payload: { nodeId: number, role: string, vcpu: number, ram: number, storage: number }) {
  const node = availableNodes.value.find(n => n.nodeId === payload.nodeId);
  if (!node) return;
  if (payload.vcpu <= 0 || payload.ram <= 0 || payload.storage <= 0) {
    addFormError.value = 'All resources must be greater than 0.';
    return;
  }
  if (payload.vcpu > getAvailableCPU(node) || payload.ram > getAvailableRAM(node) || payload.storage > getAvailableStorage(node)) {
    addFormError.value = 'Requested resources exceed available.';
    return;
  }
  addNodeLoading.value = true;
  addFormError.value = '';
  try {
    if (!cluster.value?.cluster?.name) throw new Error('Cluster name missing');
    await clusterStore.addNodesToCluster(cluster.value.cluster.name, {
      nodes: [{
        nodeId: node.nodeId,
        role: payload.role,
        vcpu: payload.vcpu,
        ram: payload.ram,
        storage: payload.storage
      }]
    });
    await clusterStore.fetchClusters();
    editNodes.value.push({ ...node, role: payload.role, vcpu: payload.vcpu, ram: payload.ram, storage: payload.storage });
    // Reset add form state
    addFormNodeId.value = null;
    addFormRole.value = 'master';
    addFormVcpu.value = 1;
    addFormRam.value = 1;
    addFormStorage.value = 1;
  } catch (e: any) {
    addFormError.value = e?.message || 'Failed to add node';
  } finally {
    addNodeLoading.value = false;
  }
}

</script>

<style scoped>
.manage-cluster-container {
  margin-top: 10rem;
  min-height: 100vh;
  background: var(--color-bg);
  padding: 0;
}
.manage-header {
  margin-bottom: var(--space-8);
}
.manage-navigation {
  display: flex;
  align-items: center;
  margin-bottom: var(--space-4);
}
.back-btn {
  color: var(--color-text-secondary) !important;
}
.back-btn:hover {
  color: var(--color-primary) !important;
  background: var(--color-primary-subtle) !important;
}
.manage-title {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text);
  margin: 0 0 var(--space-2) 0;
}
.manage-subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-secondary);
  margin: 0;
}
.status-text {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-medium);
  color: var(--color-text);
}
.status-actions {
  display: flex;
  gap: var(--space-3);
}
.status-actions.align-end {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}
.main-content-card {
  padding: unset !important;
  overflow: hidden;
}
.tab-content {
  padding: var(--space-10) var(--space-8) var(--space-8) var(--space-8);
}
.overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: var(--space-8);
}
.overview-card {
  height: 100%;
}
.details-card {
  grid-column: 1 / -1;
}
.resource-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}
.resource-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-2) 0;
  border-bottom: 1px solid var(--color-border);
}
.resource-item:last-child {
  border-bottom: none;
}
.resource-label {
  color: var(--color-text-muted);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-base);
}
.resource-value {
  color: var(--color-text);
  font-weight: var(--font-weight-semibold);
  font-family: 'Inter', monospace;
  font-size: var(--font-size-base);
}
.usage-metrics {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
.usage-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.usage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.usage-label {
  color: var(--color-text-muted);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-base);
}
.usage-value {
  color: var(--color-primary);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-base);
}
.usage-bar {
  border-radius: var(--radius-sm);
}
.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--space-4);
}
.detail-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  padding: var(--space-3);
  background: var(--color-primary-subtle);
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-md);
}
.detail-label {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}
.detail-value {
  color: var(--color-text);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-base);
}
.font-mono {
  font-family: 'Inter', monospace;
  font-size: var(--font-size-sm);
}
.truncate-cell {
  display: inline-flex;
  align-items: center;
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
}
.v-dialog .v-card-text pre {
  color: #b4befe;
  font-family: 'JetBrains Mono', 'Fira Mono', 'Menlo', 'Consolas', monospace;
  font-size: 1rem;
  line-height: 1.6;
  padding: 1.5rem;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  overflow-x: auto;
  overflow-y: auto;
  max-height: 60vh;
  box-sizing: border-box;
}
.v-dialog .v-card-text {
  padding: 0;
}
</style>
