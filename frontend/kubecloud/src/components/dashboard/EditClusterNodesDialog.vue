<template>
  <v-dialog v-model="dialog" max-width="900">
    <BaseDialogCard>
      <template #title>
        Edit Cluster Nodes
      </template>
      <template #content>
        <v-tabs v-model="editTab" class="mb-4">
          <v-tab value="list">Node List</v-tab>
          <v-tab value="add">Add Node</v-tab>
        </v-tabs>
        <div v-if="editTab === 'list'">
          <v-table v-if="editNodesWithStorage.length">
            <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>CPU</th>
                <th>RAM</th>
                <th>Storage</th>
                <th>IP</th>
                <th>Contract ID</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="node in editNodesWithStorage" :key="node.name">
                <td>{{ node.original_name }}</td>
                <td>{{ node.type }}</td>
                <td>{{ node.cpu ?? node.vcpu }}</td>
                <td>{{ node.memory ?? node.ram }} MB</td>
                <td>{{ node.storage }} MB</td>
                <td>{{ node.ip || '-' }}</td>
                <td>{{ node.contract_id || '-' }}</td>
                <td>
                  <v-btn @click="removeNode(node.original_name)"><v-icon>mdi-delete</v-icon></v-btn>
                </td>
              </tr>
            </tbody>
          </v-table>
          <div v-else class="empty-list">No nodes in this cluster.</div>
        </div>
        <div v-else-if="editTab === 'add'">
          <div class="add-form-wrapper">
            <v-text-field v-model="addFormName" label="Name" class="polished-input" />
            <v-text-field v-model.number="addFormVcpu" label="vCPU" type="number" min="1" class="polished-input" />
            <v-text-field v-model.number="addFormRam" label="RAM (MB)" type="number" min="1" class="polished-input" />
            <v-text-field v-model.number="addFormStorage" label="Storage (MB)" type="number" min="1" class="polished-input" />
            <v-select
              v-model="addFormNodeId"
              :items="availableNodesWithName"
              item-title="name"
              item-value="nodeId"
              label="Select Node"
              class="polished-input"
            >
              <template #item="{ item, props }">
                <div class="node-option-row" v-bind="props">
                  <div class="node-id">Node {{ item.raw.nodeId }}</div>
                  <div class="chip-row">
                    <v-chip color="primary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-cpu-64-bit</v-icon>
                      {{ getNodeAvailableCPU(item.raw) }} vCPU
                    </v-chip>
                    <v-chip color="success" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-memory</v-icon>
                      {{ getNodeAvailableRAM(item.raw) }} MB RAM
                    </v-chip>
                    <v-chip color="info" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-harddisk</v-icon>
                      {{ getNodeAvailableStorage(item.raw) }} MB Disk
                    </v-chip>
                    <v-chip v-if="item.raw.gpu" color="deep-purple-accent-2" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-nvidia</v-icon>
                      GPU
                    </v-chip>
                    <v-chip color="secondary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      {{ item.raw.country }}
                    </v-chip>
                  </div>
                </div>
              </template>
              <template #selection="{ item }">
                <div class="node-option-row">
                  <div class="node-id">Node {{ item.raw.nodeId }}</div>
                  <div class="chip-row">
                    <v-chip color="primary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-cpu-64-bit</v-icon>
                      {{ getNodeAvailableCPU(item.raw) }} vCPU
                    </v-chip>
                    <v-chip color="success" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-memory</v-icon>
                      {{ getNodeAvailableRAM(item.raw) }} MB RAM
                    </v-chip>
                    <v-chip color="info" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-harddisk</v-icon>
                      {{ getNodeAvailableStorage(item.raw) }} MB Disk
                    </v-chip>
                    <v-chip v-if="item.raw.gpu" color="deep-purple-accent-2" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-nvidia</v-icon>
                      GPU
                    </v-chip>
                    <v-chip color="secondary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      {{ item.raw.country }}
                    </v-chip>
                  </div>
                </div>
              </template>
            </v-select>
            <v-select v-model="addFormRole" :items="['master', 'worker']" label="Role" class="polished-input" />
            <div class="ssh-key-section" style="margin-top: 1.5rem; width: 100%;">
              <label class="ssh-key-label">SSH Key</label>
              <div v-if="sshKeysLoading" class="mb-2"><v-progress-circular indeterminate size="24" color="primary" /></div>
              <v-chip-group
                v-else
                v-model="addFormSshKey"
                :multiple="false"
                column
                class="mb-2"
              >
                <v-chip
                  v-for="key in sshKeys"
                  :key="key.ID"
                  :value="key.ID"
                  color="primary"
                  class="ma-1"
                  variant="elevated"
                >
                  {{ key.name }}
                </v-chip>
              </v-chip-group>
              <div v-if="!sshKeysLoading && sshKeys.length === 0" class="ssh-alert">
                <v-icon color="error" class="mr-1">mdi-alert-circle</v-icon>
                <span>No SSH keys found. Please add one in your dashboard.</span>
              </div>
            </div>
            <div v-if="addFormError" class="polished-error">{{ addFormError }}</div>
          </div>
        </div>
      </template>
      <template #actions>
        <div v-if="editTab === 'add'" class="add-form-actions">
          <v-btn color="primary" :loading="addNodeLoading" :disabled="!canAssignToNode || addNodeLoading" @click="confirmAddForm" class="add-node-btn">Add Node</v-btn>
          <v-btn variant="text" @click="editTab = 'list'">Cancel</v-btn>
        </div>
      </template>
    </BaseDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { getAvailableCPU, getAvailableRAM, getAvailableStorage } from '../../utils/nodeNormalizer';
import type { RentedNode } from '../../composables/useNodeManagement';
import BaseDialogCard from './BaseDialogCard.vue';
import { userService } from '../../utils/userService';

const props = defineProps<{
  modelValue: boolean,
  cluster: any,
  nodes: any[],
  loading: boolean,
  availableNodes: RentedNode[],
  addFormError: string,
  addFormNode: RentedNode | undefined,
  canAssignToNode: boolean,
  addNodeLoading: boolean,
  availableSshKeys: any[]
}>();
const emit = defineEmits(['update:modelValue', 'add-node', 'nodes-updated', 'remove-node']);
const dialog = computed({
  get: () => props.modelValue,
  set: (val: boolean) => emit('update:modelValue', val)
});
const editTab = ref('list');
const editNodes = ref<any[]>(props.nodes || []);
watch(() => props.nodes, (val) => { editNodes.value = val || []; });
const editNodesWithStorage = computed(() =>
  (editNodes.value || []).map(node => ({
    ...node,
    storage: (node.root_size || 0) + (node.disk_size || 0) + (node.storage || 0)
  }))
);
function removeNode(nodeName: string) {
  emit('remove-node', nodeName);
}
// Add node form state
const addFormNodeId = ref<number|null>(null);
const addFormRole = ref('master');
const addFormVcpu = ref(1);
const addFormRam = ref(1);
const addFormStorage = ref(1);
const addFormError = ref('');
const addFormName = ref('');
const addFormSshKey = ref<number|null>(null);
const sshKeys = ref<any[]>([]);
const sshKeysLoading = ref(false);
const sshKeysError = ref('');

onMounted(async () => {
  sshKeysLoading.value = true;
  try {
    sshKeys.value = await userService.listSshKeys();
    if (sshKeys.value.length > 0) {
      addFormSshKey.value = sshKeys.value[0].ID;
    }
  } catch (e: any) {
    sshKeysError.value = e?.message || 'Failed to load SSH keys';
  } finally {
    sshKeysLoading.value = false;
  }
});
const addFormNode = computed<RentedNode | undefined>(() => (props.availableNodes || []).find((n: RentedNode) => n.nodeId === addFormNodeId.value));
const canAssignToNode = computed(() => {
  const node = addFormNode.value;
  if (!node) return false;
  return (
    addFormVcpu.value > 0 &&
    addFormRam.value > 0 &&
    addFormStorage.value > 0 &&
    addFormVcpu.value <= getNodeAvailableCPU(node) &&
    addFormRam.value <= getNodeAvailableRAM(node) &&
    addFormStorage.value <= getNodeAvailableStorage(node)
  );
});
watch([addFormNodeId, addFormVcpu, addFormRam, addFormStorage], () => {
  const node = addFormNode.value;
  if (!node) {
    addFormError.value = '';
    return;
  }
  if (
    addFormVcpu.value > getNodeAvailableCPU(node) ||
    addFormRam.value > getNodeAvailableRAM(node) ||
    addFormStorage.value > getNodeAvailableStorage(node)
  ) {
    addFormError.value = 'Requested resources exceed available for the selected node.';
  } else {
    addFormError.value = '';
  }
});
function confirmAddForm() {
  // Find selected SSH key object
  const sshKeyObj = (sshKeys.value || []).find((k: any) => k.ID === addFormSshKey.value);
  emit('add-node', {
    name: props.cluster.cluster.name,
    token: '', // If you have a token, use it here
    nodes: [
      {
        name: addFormName.value,
        type: addFormRole.value, // 'master' | 'worker'
        node_id: addFormNodeId.value, // backend expects node_id
        cpu: addFormVcpu.value,
        memory: addFormRam.value, // already MB
        root_size: 2, // MB
        disk_size: addFormStorage.value, // already MB
        env_vars: sshKeyObj ? { SSH_KEY: sshKeyObj.public_key } : {},
      }
    ]
  });
  editTab.value = 'list';
}
// Helper to get resources already assigned to a node in this cluster
function getClusterUsedResources(nodeId: number) {
  // All values in MB
  return (editNodes.value || []).filter((n: any) => n.node_id === nodeId).reduce((acc: { cpu: number, memory: number, storage: number }, n: any) => {
    acc.cpu += n.cpu || 0;
    acc.memory += n.memory || 0; // already MB
    acc.storage += (n.root_size || 0) + (n.disk_size || 0); // already MB
    return acc;
  }, { cpu: 0, memory: 0, storage: 0 });
}
function getNodeAvailableCPU(node: RentedNode) {
  return Math.max(getAvailableCPU(node) - getClusterUsedResources(node.nodeId).cpu, 0);
}
function getNodeAvailableRAM(node: RentedNode) {
  // getAvailableRAM returns GB, convert to MB
  return Math.max((getAvailableRAM(node) * 1024) - getClusterUsedResources(node.nodeId).memory, 0);
}
function getNodeAvailableStorage(node: RentedNode) {
  // getAvailableStorage returns GB, convert to MB
  return Math.max((getAvailableStorage(node) * 1024) - getClusterUsedResources(node.nodeId).storage, 0);
}
// Ensure every node has a 'name' property for v-select display
const availableNodesWithName = computed(() =>
  (props.availableNodes || []).map(n => ({
    ...n,
    name: (n as any).name || `Node ${n.nodeId}`
  }) as RentedNode & { name: string })
);
</script>
