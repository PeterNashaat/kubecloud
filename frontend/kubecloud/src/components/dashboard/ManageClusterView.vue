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
    <v-dialog v-model="kubeconfigDialog" max-width="600">
      <v-card>
        <v-card-title>Kubeconfig</v-card-title>
        <v-card-text>
          <v-alert v-if="kubeconfigError" type="error" class="mb-4">{{ kubeconfigError }}</v-alert>
          <v-progress-linear v-if="kubeconfigLoading" indeterminate color="primary" class="mb-4" />
          <pre v-if="kubeconfigContent">{{ kubeconfigContent }}</pre>
        </v-card-text>
        <v-card-actions>
          <v-btn color="primary" @click="copyKubeconfig">Copy</v-btn>
          <v-btn color="primary" @click="downloadKubeconfigFile">Download</v-btn>
          <v-btn color="primary" @click="kubeconfigDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Modal -->
    <v-dialog v-model="showDeleteModal" max-width="400">
      <v-card>
        <v-card-title>Confirm Delete</v-card-title>
        <v-card-text>Are you sure you want to delete this cluster?</v-card-text>
        <v-card-actions>
          <v-btn color="primary" @click="showDeleteModal = false" :disabled="deletingCluster">Cancel</v-btn>
          <v-btn color="primary" @click="confirmDelete" :loading="deletingCluster" :disabled="deletingCluster">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useClusterStore } from '../../stores/clusters'
import { api } from '../../utils/api'
import { useNotificationStore } from '../../stores/notifications'

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
    console.log({data});
    
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

function formatDate(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
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

.breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.breadcrumb-item {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
}

.breadcrumb-item.active {
  color: var(--color-primary);
  font-weight: var(--font-weight-medium);
}

.breadcrumb-separator {
  color: var(--color-text-muted) !important;
  font-size: var(--font-size-sm) !important;
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

.status-bar {
  margin-bottom: var(--space-6);
}

.status-bar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.status-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.status-dot.running {
  background: var(--color-success);
}

.status-dot.stopped {
  background: var(--color-error);
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
</style>
