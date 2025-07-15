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
          <div class="main-content-card modern-cluster-card">
            <div class="modern-cluster-info">
              <div class="cluster-info-grid">
                <div class="info-label">Project Name</div>
                <div>{{ cluster.project_name || '-' }}</div>
                <div class="info-label">vCPU</div>
                <div>{{ totalVcpu }}</div>
                <div class="info-label">Created</div>
                <div>{{ formatDate(cluster.created_at) }}</div>
                <div class="info-label">Storage</div>
                <div>{{ totalStorage }} MB</div>
                <div class="info-label">Last Updated</div>
                <div>{{ formatDate(cluster.updated_at) }}</div>

                <div class="info-label">RAM</div>
                <div>{{ totalRam }} MB</div>
              </div>
            </div>
            <div class="nodes-section mt-8">
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
          </div>
        </div>
      </v-container>
    </div>

    <!-- Kubeconfig Modal -->
    <component :is="KubeconfigDialog"
      v-model="kubeconfigDialog"
      :projectName="projectName"
      :loading="kubeconfigLoading"
      :error="kubeconfigError"
      :content="kubeconfigContent"
      @copy="copyKubeconfig"
      @download="downloadKubeconfigFile"
    />

    <!-- Delete Confirmation Modal -->
    <component :is="DeleteClusterDialog"
      v-model="showDeleteModal"
      :loading="deletingCluster"
      @confirm="confirmDelete"
    />

    <!-- Edit Cluster Nodes Modal -->
    <component :is="EditClusterNodesDialog"
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
      @remove-node="handleRemoveNode"
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
import { useDeploymentEvents } from '../../composables/useDeploymentEvents';

import { getAvailableCPU, getAvailableRAM, getAvailableStorage } from '../../utils/nodeNormalizer';

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

const editClusterNodesDialog = ref(false)

// Dummy state for masters/workers (replace with real cluster data)
const editNodes = ref<any[]>([])

async function openEditClusterNodesDialog() {
  const nodesRaw = cluster.value?.cluster?.nodes;
  const nodes = Array.isArray(nodesRaw) ? nodesRaw : [];
  editNodes.value = nodes.map(n => ({ ...n }));
  editClusterNodesDialog.value = true;
  await fetchRentedNodes();
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

const { rentedNodes, loading: nodesLoading, fetchRentedNodes, addNodeToDeployment, removeNodeFromDeployment } = useNodeManagement()
const { onTaskEvent } = useDeploymentEvents();

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
    // Build the node object as expected by the backend
    const sshKey = cluster.value.cluster.env_vars?.SSH_KEY || '';
    const k3sToken = cluster.value.cluster.env_vars?.K3S_TOKEN || cluster.value.cluster.token || '';
    const nodeName = `node${Date.now()}`;
    // Convert RAM and storage to MB and enforce minimums
    const memoryMB = Math.max(Math.round(payload.ram * 1024), 256); // RAM in GB to MB, min 256MB
    const diskSizeMB = Math.max(Math.round(payload.storage * 1024), 1024); // Storage in GB to MB, min 1GB
    const rootSizeMB = 10240; // You may want to make this user-configurable
    const nodeObj = {
      name: nodeName,
      type: payload.role,
      node_id: payload.nodeId,
      cpu: payload.vcpu,
      memory: memoryMB,
      root_size: rootSizeMB,
      disk_size: diskSizeMB,
      env_vars: {
        SSH_KEY: sshKey,
        K3S_TOKEN: k3sToken
      }
    };
    const clusterPayload = {
      name: cluster.value.cluster.name,
      nodes: [nodeObj]
    };
    const response = await addNodeToDeployment(cluster.value.cluster.name, clusterPayload);
    const taskId = response.data.task_id;
    if (!taskId) throw new Error('No task ID returned from backend');
    // Wait for the deployment event for this task
    await new Promise<void>((resolve, reject) => {
      const unsubscribe = onTaskEvent(taskId, async (event: any) => {
        const status = event.data?.status || event.data?.Status || event.status;
        if (status === 'completed' || status === 'success') {
          await fetchRentedNodes();
          await clusterStore.fetchClusters();
          // Update editNodes with the latest nodes from the refreshed cluster
          const updatedCluster = clusterStore.clusters.find(c => c.project_name === cluster.value?.project_name);
          editNodes.value = updatedCluster?.cluster?.nodes && Array.isArray(updatedCluster.cluster.nodes) ? updatedCluster.cluster.nodes.map((n: any) => ({ ...n })) : [];
          // Reset add form state
          addFormNodeId.value = null;
          addFormRole.value = 'master';
          addFormVcpu.value = 1;
          addFormRam.value = 1;
          addFormStorage.value = 1;
          addNodeLoading.value = false;
          unsubscribe();
          resolve();
        } else if (status === 'failed' || status === 'error') {
          addFormError.value = event.data?.message || event.message || 'Failed to add node';
          addNodeLoading.value = false;
          unsubscribe();
          reject(new Error(addFormError.value));
        }
      });
    });
  } catch (e: any) {
    addFormError.value = e?.message || 'Failed to add node';
    addNodeLoading.value = false;
  }
}

async function handleRemoveNode(nodeName: string) {
  if (!cluster.value?.cluster?.name) return;
  try {
    await removeNodeFromDeployment(cluster.value.cluster.name, nodeName);
    await fetchRentedNodes();
    await clusterStore.fetchClusters();
    // Update editNodes with the latest nodes from the refreshed cluster
    const updatedCluster = clusterStore.clusters.find(c => c.project_name === cluster.value?.project_name);
    editNodes.value = updatedCluster?.cluster?.nodes && Array.isArray(updatedCluster.cluster.nodes) ? updatedCluster.cluster.nodes.map((n: any) => ({ ...n })) : [];
  } catch (e: any) {
    // Optionally show notification
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
.modern-cluster-card {
  background: rgba(255,255,255,0.03);
  border-radius: 0.25rem;
  padding: 2.5rem 2.5rem 2rem 2.5rem !important;
}
.modern-cluster-info {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  margin-bottom: 2.5rem;
}
.cluster-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: 1.5rem;
}
.cluster-info-grid {
  display: grid;
  grid-template-columns: 1fr 1.5fr 1fr 1.5fr;
  gap: 0.7rem 2.5rem;
  align-items: center;
}
.info-label {
  color: var(--color-text-muted);
  font-size: 1rem;
  font-weight: 500;
  text-align: right;
}
.nodes-section {
  margin-top: 2rem;
}
.dashboard-card-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-text);
  display: flex;
  align-items: center;
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
.empty-message {
  color: var(--color-text-muted);
  text-align: center;
  margin: 2rem 0;
}
@media (max-width: 900px) {
  .cluster-info-grid {
    grid-template-columns: 1fr 1.5fr;
  }
}
@media (max-width: 600px) {
  .modern-cluster-card {
    padding: 1.2rem 0.5rem 1rem 0.5rem;
  }
  .cluster-title {
    font-size: 1.3rem;
  }
  .cluster-info-grid {
    grid-template-columns: 1fr;
    gap: 0.5rem 1rem;
  }
  .info-label {
    font-size: 1rem;
    text-align: left;
  }
}
</style>
