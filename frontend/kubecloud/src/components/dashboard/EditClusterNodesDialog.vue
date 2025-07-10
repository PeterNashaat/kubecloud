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
                <td>{{ node.name }}</td>
                <td>{{ node.type }}</td>
                <td>{{ node.cpu ?? node.vcpu }}</td>
                <td>{{ node.memory ?? node.ram }} MB</td>
                <td>{{ node.storage }} MB</td>
                <td>{{ node.ip || '-' }}</td>
                <td>{{ node.contract_id || '-' }}</td>
                <td>
                  <v-btn @click="removeNode(node.name)"><v-icon>mdi-delete</v-icon></v-btn>
                </td>
              </tr>
            </tbody>
          </v-table>
          <div v-else class="empty-list">No nodes in this cluster.</div>
        </div>
        <div v-else-if="editTab === 'add'">
          <div class="add-form-wrapper">
            <v-select
              v-model="addFormNodeId"
              :items="availableNodes"
              item-title="name"
              item-value="nodeId"
              label="Select Node"
              :disabled="loading"
              :return-object="false"
              :item-props="nodeDropdownProps"
              class="polished-input"
              :menu-props="{ maxHeight: '300px' }"
            >
              <template #item="{ item }">
                <div class="node-dropdown-item">
                  <div class="chip-row">
                    <span class="spec-chip">vCPU: {{ getAvailableCPU(item.raw) }}</span>
                    <span class="spec-chip">RAM: {{ getAvailableRAM(item.raw) }} MB</span>
                    <span class="spec-chip">Storage: {{ getAvailableStorage(item.raw) }} MB</span>
                    <span v-if="item.raw.gpu" class="spec-chip">GPU: {{ item.raw.gpu }}</span>
                    <span v-if="item.raw.country" class="spec-chip">{{ item.raw.country }}</span>
                  </div>
                  <span class="node-id">ID: {{ item.raw.nodeId }}</span>
                </div>
              </template>
            </v-select>
            <v-text-field v-model.number="addFormVcpu" label="vCPU" type="number" min="1" class="polished-input" />
            <v-text-field v-model.number="addFormRam" label="RAM (MB)" type="number" min="1" class="polished-input" />
            <v-text-field v-model.number="addFormStorage" label="Storage (MB)" type="number" min="1" class="polished-input" />
            <v-select v-model="addFormRole" :items="['master', 'worker']" label="Role" class="polished-input" />
            <div v-if="addFormError" class="polished-error">{{ addFormError }}</div>
          </div>
        </div>
      </template>
      <template #actions>
        <div v-if="editTab === 'add'" class="add-form-actions">
          <v-btn color="primary" :loading="addNodeLoading" :disabled="!canAssignToNode || addNodeLoading" @click="confirmAddForm">Add Node</v-btn>
          <v-btn variant="text" @click="editTab = 'list'">Cancel</v-btn>
        </div>
      </template>
    </BaseDialogCard>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { getAvailableCPU, getAvailableRAM, getAvailableStorage } from '../../utils/nodeNormalizer';
import type { RentedNode } from '../../composables/useNodeManagement';
import BaseDialogCard from './BaseDialogCard.vue';

const props = defineProps<{
  modelValue: boolean,
  cluster: any,
  nodes: any[],
  loading: boolean,
  availableNodes: RentedNode[],
  addFormError: string,
  addFormNode: RentedNode | undefined,
  canAssignToNode: boolean,
  addNodeLoading: boolean
}>();
const emit = defineEmits(['update:modelValue', 'add-node', 'nodes-updated']);
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
function closeDialog() { emit('update:modelValue', false); }
function removeNode(nodeName: string) {
  emit('nodes-updated', editNodes.value.filter((n: any) => n.name !== nodeName));
}
// Add node form state
const addFormNodeId = ref<number|null>(null);
const addFormRole = ref('master');
const addFormVcpu = ref(1);
const addFormRam = ref(1);
const addFormStorage = ref(1);
const addFormError = ref('');
const addFormNode = computed<RentedNode | undefined>(() => (props.availableNodes || []).find((n: RentedNode) => n.nodeId === addFormNodeId.value));
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
function confirmAddForm() {
  emit('add-node', {
    nodeId: addFormNodeId.value,
    role: addFormRole.value,
    vcpu: addFormVcpu.value,
    ram: addFormRam.value,
    storage: addFormStorage.value
  });
  editTab.value = 'list';
}
function nodeDropdownProps(node: any) {
  return {
    title: node.name,
    subtitle: `vCPU: ${getAvailableCPU(node)}, RAM: ${getAvailableRAM(node)} MB, Storage: ${getAvailableStorage(node)} MB`,
  };
}
</script>

<style scoped>
.add-form-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem 1rem 1rem 1rem;
  background: var(--color-surface-2, #23243a);
  border-radius: 14px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  min-width: 350px;
  max-width: 500px;
  margin: 0 auto;
}
.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.spec-chip {
  background: var(--color-surface-2, #23243a);
  color: var(--color-text, #cfd2fa);
  border-radius: 8px;
  padding: 0.2rem 0.7rem;
  font-size: 0.95rem;
}
.node-id {
  font-weight: 600;
  color: var(--color-text);
  font-size: 0.98em;
}
.node-dropdown-item {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  padding: 0.2rem 0.5rem;
}
.polished-input {
  width: 100%;
  margin-bottom: 0.7rem;
}
.add-form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 0.5rem;
  width: 100%;
}
.polished-error {
  font-size: 1.05rem;
  font-weight: 500;
  color: #ff6b6b !important;
  background: #2d1a1a !important;
  border-radius: 8px;
  margin-bottom: 0.7rem;
}
.empty-list {
  color: #888;
  font-style: italic;
  margin-bottom: 0.5rem;
}
</style> 