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
            <h1 class="manage-title">{{ cluster.name }}</h1>
            <p class="manage-subtitle">Manage your Kubernetes cluster configuration and resources</p>
            <div class="tag-list mt-2">
              <v-chip v-for="tag in cluster.tags" :key="tag" size="x-small" class="mr-1">{{ tag }}</v-chip>
            </div>
          </div>
        </div>
        <div v-if="!loading && !notFound && cluster" class="manage-content-wrapper">
          <!-- Status Bar & Actions -->
          <div class="card status-bar">
            <div class="status-bar-content">
              <div class="status-indicator">
                <span class="status-dot" :class="cluster.status === 'running' ? 'running' : 'stopped'"></span>
                <span class="status-text">{{ cluster.status }}</span>
              </div>
              <div class="status-actions">
                <v-btn variant="outlined" class="btn btn-outline" @click="downloadKubeconfig">
                  <v-icon icon="mdi-download" class="mr-2"></v-icon>
                  Download Kubeconfig
                </v-btn>
                <v-btn variant="outlined" class="btn btn-outline" @click="openDashboard">
                  <v-icon icon="mdi-view-dashboard" class="mr-2"></v-icon>
                  Open Dashboard
                </v-btn>
                <v-btn variant="outlined" class="btn btn-outline" color="success" v-if="cluster.status !== 'running'" @click="startCluster">
                  <v-icon icon="mdi-play" class="mr-2"></v-icon>
                  Start
                </v-btn>
                <v-btn variant="outlined" class="btn btn-outline" color="warning" v-if="cluster.status === 'running'" @click="stopCluster">
                  <v-icon icon="mdi-stop" class="mr-2"></v-icon>
                  Stop
                </v-btn>
                <v-btn variant="outlined" class="btn btn-outline" color="error" @click="deleteCluster">
                  <v-icon icon="mdi-delete" class="mr-2"></v-icon>
                  Delete
                </v-btn>
              </div>
            </div>
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
                        <span class="resource-value">{{ cluster.nodes }}</span>
                      </div>
                      <div class="resource-item">
                        <span class="resource-label">vCPU:</span>
                        <span class="resource-value">{{ cluster.cpu }}</span>
                      </div>
                      <div class="resource-item">
                        <span class="resource-label">RAM:</span>
                        <span class="resource-value">{{ cluster.memory }}</span>
                      </div>
                      <div class="resource-item">
                        <span class="resource-label">Storage:</span>
                        <span class="resource-value">{{ cluster.storage }}</span>
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
                        <span class="detail-label">Region:</span>
                        <span class="detail-value">{{ cluster.region }}</span>
                      </div>
                      <div class="detail-item">
                        <span class="detail-label">Created:</span>
                        <span class="detail-value">{{ cluster.createdAt }}</span>
                      </div>
                      <div class="detail-item">
                        <span class="detail-label">Est. Cost:</span>
                        <span class="detail-value">${{ cluster.cost }}/month</span>
                      </div>
                      <div class="detail-item">
                        <span class="detail-label">Last Updated:</span>
                        <span class="detail-value">{{ cluster.lastUpdated }}</span>
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
                <v-table v-if="clusterNodes.length">
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Type</th>
                      <th>CPU</th>
                      <th>RAM</th>
                      <th>Storage</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="node in clusterNodes" :key="node.Name">
                      <td>{{ node.Name }}</td>
                      <td>{{ node.Type }}</td>
                      <td>{{ node.CPU }}</td>
                      <td>{{ node.Memory }} MB</td>
                      <td>{{ node.RootSize + node.DiskSize }} MB</td>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useClusterStore } from '../../stores/clusters'

const router = useRouter()
const route = useRoute()
const clusterStore = useClusterStore()

const loading = ref(true)
const notFound = ref(false)

const clusterId = computed(() => route.params.id?.toString() || '')
const cluster = computed(() =>
  clusterStore.clusters.find(c => c.id === clusterId.value)
)

const tab = ref('overview')

// Node details (mocked or fetched from backend if available)
const clusterNodes = ref<any[]>([])

// Metrics
const metrics = ref<any>(null)
const metricsLoading = ref(false)
const metricsError = ref<string | null>(null)

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)
const eventsError = ref<string | null>(null)

const loadCluster = async () => {
  loading.value = true
  notFound.value = false
  try {
    if (!clusterStore.clusters.length) {
      await clusterStore.fetchClusters()
    }
    if (!cluster.value) {
      notFound.value = true
    } else {
      // Optionally fetch node details here
      // clusterNodes.value = await api.get(`/clusters/${clusterId.value}/nodes`)
      // For now, mock nodes if needed
      clusterNodes.value = [
        { Name: 'master-1', Type: 'master', CPU: 2, Memory: 4096, RootSize: 10240, DiskSize: 10240 },
        { Name: 'worker-1', Type: 'worker', CPU: 2, Memory: 4096, RootSize: 10240, DiskSize: 10240 },
      ]
      // Fetch metrics
      fetchMetrics()
      // Fetch events
      fetchEvents()
    }
  } catch (e) {
    notFound.value = true
  } finally {
    loading.value = false
  }
}

onMounted(loadCluster)
watch(() => clusterId.value, loadCluster)

const goBack = () => {
  router.push('/dashboard')
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

// Actions
function downloadKubeconfig() {
  // TODO: Implement download logic
  // eslint-disable-next-line no-console
  console.log('Download kubeconfig')
}
function openDashboard() {
  // TODO: Implement open dashboard logic
  // eslint-disable-next-line no-console
  console.log('Open dashboard')
}
function startCluster() {
  // TODO: Implement start logic
  // eslint-disable-next-line no-console
  console.log('Start cluster')
}
function stopCluster() {
  // TODO: Implement stop logic
  // eslint-disable-next-line no-console
  console.log('Stop cluster')
}
function deleteCluster() {
  // TODO: Implement delete logic
  // eslint-disable-next-line no-console
  console.log('Delete cluster')
}

// Metrics fetching
async function fetchMetrics() {
  metricsLoading.value = true
  metricsError.value = null
  try {
    metrics.value = await clusterStore.getClusterMetrics(clusterId.value)
  } catch (e: any) {
    metricsError.value = e.message || 'Failed to fetch metrics'
  } finally {
    metricsLoading.value = false
  }
}

// Events fetching (mocked for now)
async function fetchEvents() {
  eventsLoading.value = true
  eventsError.value = null
  try {
    // TODO: Replace with real API call
    events.value = [
      { id: 1, message: 'Cluster created', timestamp: new Date().toISOString() },
      { id: 2, message: 'Node master-1 started', timestamp: new Date().toISOString() },
      { id: 3, message: 'Node worker-1 started', timestamp: new Date().toISOString() },
    ]
  } catch (e: any) {
    eventsError.value = e.message || 'Failed to fetch events'
  } finally {
    eventsLoading.value = false
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

@media (max-width: 1100px) {
  .overview-grid {
    grid-template-columns: 1fr;
  }
  .details-card {
    grid-column: auto;
  }
}

@media (max-width: 768px) {
  .manage-cluster-container {
    padding: var(--space-4);
  }
  
  .manage-title {
    font-size: var(--font-size-2xl);
  }
  
  .manage-subtitle {
    font-size: var(--font-size-base);
  }
  
  .status-bar-content {
    flex-direction: column;
    gap: var(--space-4);
    align-items: flex-start;
  }
  
  .status-actions {
    width: 100%;
    justify-content: flex-start;
  }
  
  .tab-content {
    padding: var(--space-6) var(--space-4) var(--space-4) var(--space-4);
  }
  
  .overview-grid {
    gap: var(--space-4);
  }
  
  .details-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 480px) {
  .manage-cluster-container {
    padding: var(--space-3);
  }
  
  .manage-title {
    font-size: var(--font-size-xl);
  }
  
  .status-actions {
    flex-direction: column;
    gap: var(--space-2);
  }
  
  .tab-content {
    padding: var(--space-4) var(--space-3) var(--space-3) var(--space-3);
  }
}
</style>
