<template>
  <div class="deploy-container">
    <v-container fluid class="pa-0">
      <div class="deploy-header mb-6">
        <h1 class="hero-title">Deploy New Cluster</h1>
        <p class="section-subtitle">Create and configure your Kubernetes cluster in just a few steps</p>
      </div>
      
      <div class="deploy-content-wrapper">
        <div class="deploy-card">
          <!-- Progress Indicator -->
          <div class="progress-section">
            <div class="stepper">
              <div class="step" :class="{ active: step >= 1, completed: step > 1 }">
                <div class="step-circle">1</div>
                <div class="step-label">Define VMs</div>
              </div>
              <div class="step" :class="{ active: step >= 2, completed: step > 2 }">
                <div class="step-circle">2</div>
                <div class="step-label">Place VMs</div>
              </div>
              <div class="step" :class="{ active: step >= 3 }">
                <div class="step-circle">3</div>
                <div class="step-label">Review</div>
              </div>
            </div>
          </div>

          <!-- Step Content -->
          <div class="step-content">
            <Step1DefineVMs
              v-if="step === 1"
              :masters="masters"
              :workers="workers"
              :availableSshKeys="availableSshKeys"
              :addMaster="addMaster"
              :addWorker="addWorker"
              :removeMaster="removeMaster"
              :removeWorker="removeWorker"
              :openEditNodeModal="openEditNodeModal"
              :selectedSshKeys="selectedSshKeys"
              :setSelectedSshKeys="setSelectedSshKeys"
              :isStep1Valid="isStep1Valid"
              @navigateToSshKeys="navigateToSshKeys"
              @nextStep="nextStep"
            />
            <Step2AssignNodes
              v-else-if="step === 2"
              :allVMs="allVMs"
              :availableNodes="availableNodes.map(n => ({ id: String(n.id), label: n.label }))"
              :getNodeInfo="getNodeInfoString"
              :onAssignNode="onAssignNode"
              :isStep2Valid="isStep2Valid"
              @nextStep="nextStep"
              @prevStep="prevStep"
            />
            <Step3Review
              v-else-if="step === 3"
              :allVMs="allVMs"
              :getNodeInfo="getNodeInfoString"
              :onDeployCluster="onDeployCluster"
              :prevStep="prevStep"
              :deploying="deploying"
            />
          </div>
        </div>
      </div>
    </v-container>
    <EditNodeModal v-if="editNodeModal.open && editNodeModal.node" :node="editNodeModal.node" :visible="editNodeModal.open" :availableSshKeys="availableSshKeys" @save="saveEditNode" @cancel="closeEditNodeModal" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useClusterStore } from '../stores/clusters';
import NodeCard from '../components/NodeCard.vue';
import EditNodeModal from '../components/EditNodeModal.vue';
import { useDeployCluster } from '../composables/useDeployCluster';
import type { VM, SshKey } from '../composables/useDeployCluster';
import Step1DefineVMs from '../components/deploy/Step1DefineVMs.vue';
import Step2AssignNodes from '../components/deploy/Step2AssignNodes.vue';
import Step3Review from '../components/deploy/Step3Review.vue';
import { api } from '../utils/api';
import { useNotificationStore } from '../stores/notifications';
import { UserService } from '../utils/userService';
import { useNodes } from '../composables/useNodes';

const notificationStore = useNotificationStore();
const userService = new UserService();

// Type definitions
interface Node {
  id: number;
  label: string;
  totalCPU: number;
  totalRAM: number;
  hasGPU: boolean;
  location: string;
}

interface ClusterPayload {
  Master: K8sNode | null;
  Workers: K8sNode[];
  Token: string;
  NetworkName: string;
  SSHKey: string;
  Flist: string;
  SolutionType: string;
}

interface K8sNode {
  VM: VMFull;
  DiskSizeGB: number;
}

interface VMFull {
  Name: string;
  NodeID: number | null;
  CPU: number;
  MemoryMB: number;
  NetworkName: string;
  Flist: string;
  Entrypoint: string;
  EnvVars: Record<string, string>;
  RootfsSizeMB: number;
  PublicIP: boolean;
  Planetary: boolean;
  QEMUArgs: any[];
}

const step = ref(1);
const { masters, workers, availableSshKeys, addMaster, addWorker, removeMaster, removeWorker } = useDeployCluster();
const deploying = ref(false);

const allVMs = computed(() => [...masters.value, ...workers.value]);

const { fetchNodes, nodes, loading: nodesLoading, error: nodesError } = useNodes();
const availableNodes = ref<any[]>([]);

async function fetchAvailableNodes() {
  try {
    await fetchNodes();
    const nodesArr = Array.isArray(nodes.value) ? nodes.value : [];
    availableNodes.value = nodesArr.map((n: any) => ({
      id: String(n.id),
      label: n.label || n.name || `Node #${n.id}`,
      totalCPU: n.resources?.cpu || n.totalCPU || 0,
      totalRAM: n.resources?.memory || n.totalRAM || 0,
      hasGPU: n.gpu || n.hasGPU || false,
      location: n.location || '',
    }));
  } catch (err) {
    availableNodes.value = [];
  }
}

// Computed properties for totals
const totalVcpu = computed(() => {
  return allVMs.value.reduce((total, vm) => total + vm.vcpu, 0);
});

const totalRam = computed(() => {
  return allVMs.value.reduce((total, vm) => total + vm.ram, 0);
});

const clusterName = ref('');
const selectedSshKeys = ref<string[]>([]);
const qsfsConfig = ref('');

// Cluster name generator words
const adjectives = [
  'swift', 'bright', 'cosmic', 'quantum', 'stellar', 'azure', 'crimson', 'golden',
  'silver', 'emerald', 'sapphire', 'crystal', 'thunder', 'lightning', 'storm',
  'ocean', 'mountain', 'forest', 'desert', 'arctic', 'tropical', 'mystic'
];

const nouns = [
  'cluster', 'cloud', 'node', 'server', 'engine', 'core', 'hub', 'nexus',
  'forge', 'vault', 'tower', 'citadel', 'fortress', 'sanctuary', 'haven',
  'realm', 'domain', 'sphere', 'matrix', 'grid', 'network', 'system'
];

// --- Step 1 Validation ---
const isStep1Valid = computed(() => {
  if (masters.value.length === 0) return false;
  // Every node (master/worker) must have at least one SSH key
  return allVMs.value.every(vm => Array.isArray(vm.sshKeyIds) && vm.sshKeyIds.length > 0);
});

// --- Step 2 Validation ---
const assignedNodeIds = computed(() => allVMs.value.map((vm: any) => vm.node));
const allVMsAssigned = computed(() => allVMs.value.length > 0 && allVMs.value.every((vm: any) => vm.node !== null && vm.node !== undefined));
const uniqueNodeAssignment = computed(() => {
  const nodeIds: number[] = assignedNodeIds.value.filter((id: number | null) => id !== null);
  return new Set(nodeIds).size === nodeIds.length;
});
const isStep2Valid = computed(() => allVMsAssigned.value && uniqueNodeAssignment.value);

// --- Step 3 Validation ---
const isStep3Valid = computed(() => isStep2Valid.value && isStep1Valid.value);

// Helper function to get node info
function getNodeInfo(nodeId: string | null) {
  if (!nodeId) return '';
  const node = availableNodes.value.find(n => n.id === nodeId);
  if (!node) return '';
  return `${node.totalCPU} vCPU, ${node.totalRAM}GB RAM${node.hasGPU ? ', GPU Available' : ''}`;
}

// --- Deploy Logic ---
const clusterToken = ref('securetoken');
const clusterNetworkName = ref('');
const defaultFlist = ref('https://hub.grid.tf/tf-official-apps/threefolddev-k3s-v1.31.0.flist');
const defaultEntrypoint = ref('/sbin/zinit init');

// Cluster name generator
function generateClusterName() {
  const randomAdjective = adjectives[Math.floor(Math.random() * adjectives.length)];
  const randomNoun = nouns[Math.floor(Math.random() * nouns.length)];
  const randomNumber = Math.floor(Math.random() * 999) + 1;
  clusterName.value = `${randomAdjective}-${randomNoun}-${randomNumber}`;
}

// Navigate to SSH keys management
function navigateToSshKeys() {
  // This would navigate to the SSH keys page in the dashboard
  // For now, we'll just show an alert
  alert('This would navigate to the SSH Keys management page in the dashboard');
}

// Get SSH key name by ID
function getSshKeyName(keyId: number) {
  const key = availableSshKeys.value.find(k => k.id === keyId);
  return key ? key.name : 'Unknown';
}

// Get validation message for form errors
function getValidationMessage() {
  const errors = [];
  
  if (!clusterName.value) {
    errors.push('Cluster name is required');
  }
  
  if (selectedSshKeys.value.length === 0) {
    errors.push('At least one SSH key must be selected');
  }
  
  if (!allVMsAssigned.value) {
    errors.push('All VMs must be assigned to nodes');
  }
  
  return errors.join('. ');
}

// --- Navigation ---
function nextStep() {
  if ((step.value === 1 && isStep1Valid.value) || (step.value === 2 && isStep2Valid.value)) {
    step.value++;
  }
}
function prevStep() {
  if (step.value > 1) step.value--;
}

const clusters = useClusterStore();

const clusterPayload = computed<ClusterPayload>(() => {
  const networkName = clusterNetworkName.value || `${clusterName.value}_network`;
  const token = clusterToken.value;
  const flist = 'https://hub.grid.tf/tf-official-apps/threefolddev-k3s-v1.31.0.flist';
  const entrypoint = '/sbin/zinit init';
  const sshKeyObj = availableSshKeys.value.find(k => k.id === selectedSshKeys.value[0]);
  const sshKey = sshKeyObj ? sshKeyObj.fingerprint : '';
  function buildVM(vm: VM, nodeType: string): VMFull {
    return {
      Name: vm.name,
      NodeID: vm.node,
      CPU: vm.vcpu,
      MemoryMB: vm.ram * 1024,
      NetworkName: networkName,
      Flist: flist,
      Entrypoint: entrypoint,
      EnvVars: {
        SSH_KEY: sshKey,
        K3S_TOKEN: token,
        K3S_DATA_DIR: '/mnt/data',
        K3S_FLANNEL_IFACE: 'eth0',
        K3S_NODE_NAME: vm.name,
        K3S_URL: '',
      },
      RootfsSizeMB: (vm.rootfs || 10) * 1024,
      PublicIP: vm.publicIp,
      Planetary: vm.planetary,
      QEMUArgs: [],
    };
  }
  const masterNode = masters.value[0];
  const master = masterNode ? {
    VM: buildVM(masterNode, 'master'),
    DiskSizeGB: masterNode.rootfs || 10,
  } : null;
  const workersArr = workers.value.map(worker => ({
    VM: buildVM(worker, 'worker'),
    DiskSizeGB: worker.rootfs || 10,
  }));
  return {
    Master: master,
    Workers: workersArr,
    Token: token,
    NetworkName: networkName,
    SSHKey: sshKey,
    Flist: flist,
    SolutionType: `kubernetes/user/${clusterName.value}`,
  };
});

async function onDeployCluster() {
  deploying.value = true;
  try {
    await api.post('/v1/deploy', clusterPayload.value, {
      showNotifications: true,
      loadingMessage: 'Deploying cluster...',
      successMessage: 'Cluster deployed successfully!',
      errorMessage: 'Failed to deploy cluster',
      requiresAuth: true
    });
    notificationStore.success('Success', 'Cluster deployed successfully!');
    // Optionally, redirect or reset wizard here
  } catch (err: any) {
    notificationStore.error('Error', err.message || 'Failed to deploy cluster');
  } finally {
    deploying.value = false;
  }
}

// Initialize component
onMounted(() => {
  // Auto-generate cluster name on component mount
  generateClusterName();
  fetchAvailableNodes();
});

const editNodeModal = ref({ open: false, type: '', idx: -1, node: null as null | VM });
function openEditNodeModal(type: 'master' | 'worker', idx: number) {
  const node = type === 'master' ? { ...masters.value[idx] } : { ...workers.value[idx] };
  editNodeModal.value = { open: true, type, idx, node };
}
function closeEditNodeModal() {
  editNodeModal.value = { open: false, type: '', idx: -1, node: null };
}
function saveEditNode(updatedNode: VM) {
  if (!editNodeModal.value.node) return;
  if (editNodeModal.value.type === 'master') {
    masters.value[editNodeModal.value.idx] = { ...updatedNode };
  } else if (editNodeModal.value.type === 'worker') {
    workers.value[editNodeModal.value.idx] = { ...updatedNode };
  }
  closeEditNodeModal();
}

const editNodeValidation = computed(() => {
  const node = editNodeModal.value.node;
  if (!node) return { valid: false };
  const errors: Record<string, string> = {};
  if (!node.name || !node.name.trim()) errors.name = 'Name is required.';
  if (!node.vcpu || node.vcpu <= 0) errors.vcpu = 'vCPU must be a positive number.';
  if (!node.ram || node.ram <= 0) errors.ram = 'RAM must be a positive number.';
  if (!node.rootfs || node.rootfs <= 0) errors.rootfs = 'Rootfs size must be positive.';
  if (!node.disk || node.disk <= 0) errors.disk = 'Disk size must be positive.';
  if (!node.sshKeyIds || node.sshKeyIds.length === 0) errors.ssh = 'At least one SSH key must be selected.';
  return { valid: Object.keys(errors).length === 0, errors };
});

function setSelectedSshKeys(keys: string[]) {
  selectedSshKeys.value = keys;
}
function onAssignNode(vmIdx: number, nodeId: string) {
  allVMs.value[vmIdx].node = nodeId ? Number(nodeId) : null;
}
function getNodeInfoString(id: string) {
  return getNodeInfo(id);
}
</script>

<style>
.deploy-container {
  /* Enhanced palette for Deploy Cluster only */
  --color-bg-elevated: #20243a;
  --color-bg-hover: #23263b;
  --color-chip-bg: #23263b;
  --color-chip-border: #334155;
  --shadow-card: 0 6px 24px rgba(16, 24, 40, 0.12);
}
</style>

<style scoped>
.deploy-container {
  min-height: 100vh;
  background: var(--color-bg, #15162b);
  padding-top: 3.5rem;
  margin-top: 7rem;
}
.deploy-header {
  text-align: center;
  margin-bottom: 2.5rem;
}
.hero-title {
  font-size: 2.2rem;
  font-weight: 700;
  color: var(--color-text, #fff);
  margin-bottom: 0.5rem;
}
.section-subtitle {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 1.1rem;
}
.deploy-content-wrapper {
  display: flex;
  justify-content: center;
}
.deploy-card {
  background: var(--color-surface-1, #18192b);
  border-radius: 22px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.12);
  padding: 3.5rem 3rem 2.5rem 3rem;
  width: 100%;
  max-width: 900px;
  margin-top: 2.5rem;
}
.progress-section {
  margin-bottom: 3rem;
}
.stepper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2.2rem;
  position: relative;
}
.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  position: relative;
  z-index: 1;
}
.step-circle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--color-surface-2, #23243a);
  color: var(--color-primary, #6366f1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1.1rem;
  margin-bottom: 0.3rem;
  border: 2px solid var(--color-surface-2, #23243a);
  transition: background 0.2s, color 0.2s, border 0.2s;
  position: relative;
  z-index: 2;
}
.step.active .step-circle {
  background: var(--color-primary, #6366f1);
  color: #fff;
  border: 2px solid var(--color-primary, #6366f1);
}
.step.completed .step-circle {
  background: var(--color-success, #22d3ee);
  color: #fff;
  border: 2px solid var(--color-success, #22d3ee);
}

.step-label {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 1rem;
  margin-top: 0.2rem;
  text-align: center;
  font-weight: 500;
  letter-spacing: 0.01em;
}
.step.active .step-label {
  color: var(--color-primary, #6366f1);
  font-weight: 600;
}
.step.completed .step-label {
  color: var(--color-success, #22d3ee);
  font-weight: 600;
}
.step:not(:last-child)::after {
  content: '';
  position: absolute;
  top: 18px;
  right: -50%;
  width: 100%;
  height: 4px;
  background: var(--color-surface-2, #23243a);
  z-index: 0;
}
.step.completed:not(:last-child)::after {
  background: var(--color-success, #22d3ee);
}
@media (max-width: 900px) {
  .deploy-card {
    padding: 1.2rem 0.5rem 1.2rem 0.5rem;
  }
  .stepper {
    flex-direction: column;
    gap: 1.2rem;
  }
  .step {
    flex-direction: row;
    align-items: center;
    gap: 0.7rem;
  }
  .step-label {
    margin-top: 0;
    margin-left: 0.7rem;
    text-align: left;
  }
}
</style>
