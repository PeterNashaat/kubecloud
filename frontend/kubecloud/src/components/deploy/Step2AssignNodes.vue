<template>
  <div>
    <div class="section-header">
      <h3 class="section-title">
        <v-icon icon="mdi-server-network" class="mr-2"></v-icon>
        Assign VMs to Reserved Nodes
      </h3>
      <p class="section-subtitle">Select which reserved nodes will host your cluster VMs</p>
    </div>
    <div class="vm-assignment-grid">
      <div v-for="(vm, index) in allVMs" :key="index" class="vm-assignment-card">
        <div class="vm-assignment-header">
          <div class="vm-avatar" :class="vm.name.includes('Master') ? 'master' : 'worker'">
            <v-icon :icon="vm.name.includes('Master') ? 'mdi-server' : 'mdi-desktop-tower'" color="white"></v-icon>
          </div>
          <div class="vm-info">
            <h4 class="vm-title">{{ vm.name }}</h4>
            <div class="vm-specs">
              <span class="spec-chip">{{ vm.vcpu }} vCPU</span>
              <span class="spec-chip">{{ vm.ram }}GB RAM</span>
            </div>
          </div>
        </div>
        <v-select
          v-model="vm.node"
          :items="availableNodes"
          :item-title="nodeLabel"
          item-value="nodeId"
          label="Select Reserved Node"
          clearable
          class="node-select"
          @update:modelValue="val => props.onAssignNode(index, val)"
        >
          <template #item="{ item, index, props: itemProps }">
            <div>
              <div v-bind="itemProps" class="node-option-row">
                <div class="node-id">Node {{ item.raw.nodeId }}</div>
                <div class="chip-row">
                  <v-chip color="primary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                    <v-icon size="14" class="mr-1">mdi-cpu-64-bit</v-icon>
                    {{ item.raw.cpu }} vCPU
                  </v-chip>
                  <v-chip color="success" text-color="white" size="x-small" class="mr-1" variant="outlined">
                    <v-icon size="14" class="mr-1">mdi-memory</v-icon>
                    {{ item.raw.ram }} GB RAM
                  </v-chip>
                  <v-chip color="info" text-color="white" size="x-small" class="mr-1" variant="outlined">
                    <v-icon size="14" class="mr-1">mdi-harddisk</v-icon>
                    {{ item.raw.storage }} GB Disk
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
              <v-divider v-if="index < availableNodes.length - 1" />
            </div>
          </template>
          <template #selection="{ item }">
            <div class="node-id">Node {{ item.raw.nodeId }}</div>

            <div class="chip-row">
              <v-chip color="primary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                <v-icon size="14" class="mr-1">mdi-cpu-64-bit</v-icon>
                {{ item.raw.cpu }} vCPU
              </v-chip>
              <v-chip color="success" text-color="white" size="x-small" class="mr-1" variant="outlined">
                <v-icon size="14" class="mr-1">mdi-memory</v-icon>
                {{ item.raw.ram }} GB RAM
              </v-chip>
              <v-chip color="info" text-color="white" size="x-small" class="mr-1" variant="outlined">
                <v-icon size="14" class="mr-1">mdi-harddisk</v-icon>
                {{ item.raw.storage }} GB Disk
              </v-chip>
              <v-chip v-if="item.raw.gpu" color="deep-purple-accent-2" text-color="white" size="x-small" class="mr-1" variant="outlined">
                <v-icon size="14" class="mr-1">mdi-nvidia</v-icon>
                GPU
              </v-chip>
              <v-chip color="secondary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                {{ item.raw.country }}
              </v-chip>
            </div>
          </template>
        </v-select>
      </div>
    </div>
    <div class="step-actions">
      <v-btn variant="outlined" @click="$emit('prevStep')">
        <v-icon start icon="mdi-arrow-left"></v-icon>
        Back
      </v-btn>
      <v-btn color="primary" :disabled="!isStep2Valid" @click="$emit('nextStep')">
        Continue
        <v-icon end icon="mdi-arrow-right"></v-icon>
      </v-btn>
    </div>
  </div>
</template>
<script setup lang="ts">
import type { VM } from '../../composables/useDeployCluster';
import { defineProps, withDefaults, defineEmits } from 'vue';
const props = withDefaults(defineProps<{
  allVMs: VM[];
  availableNodes: {
    nodeId: number;
    cpu: number;
    ram: number;
    storage: number;
    price_usd: number | null;
    gpu: boolean;
    locationString: string;
    country: string;
    city: string;
    status: string;
    healthy: boolean;
    rentable: boolean;
    rented: boolean;
    dedicated: boolean;
    certificationType: string;
    [key: string]: any;
  }[];
  getNodeInfo: (id: number) => string;
  onAssignNode: (vmIdx: number, nodeId: number) => void;
  isStep2Valid?: boolean;
}>(), {
  isStep2Valid: false
});
const emit = defineEmits(['nextStep', 'prevStep']);

function nodeLabel(node: any) {
  if (!node) return '';
  return `Node ${node.nodeId}`;
}
</script>
<script lang="ts">
export default {
  name: 'Step2AssignNodes'
};
</script>
<style scoped>
.section-header {
  margin-bottom: 2rem;
}
.section-title {
  font-size: 1.2rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.section-subtitle {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 1rem;
  margin-top: 0.25rem;
}
.vm-assignment-grid {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
}
.vm-assignment-card {
  background: var(--color-surface-1, #18192b);
  border-radius: 12px;
  padding: 1.5rem;
  margin-bottom: 2rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
  flex: 1 1 320px;
  min-width: 320px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.vm-assignment-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}
.vm-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary, #6366f1);
}
.vm-avatar.worker {
  background: var(--color-success, #22d3ee);
}
.vm-title {
  font-size: 1.1rem;
  font-weight: 500;
}
.vm-specs {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.25rem;
}
.spec-chip {
  background: var(--color-surface-2, #23243a);
  color: var(--color-text, #cfd2fa);
  border-radius: 8px;
  padding: 0.2rem 0.7rem;
  font-size: 0.95rem;
}
.node-select {
  margin-top: 0.5rem;
}
.step-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 2rem;
}
.btn-outline {
  border: 1px solid var(--color-primary, #6366f1);
  color: var(--color-primary, #6366f1);
}
.btn-primary {
  background: var(--color-primary, #6366f1);
  color: #fff;
}
@media (max-width: 900px) {
  .vm-assignment-grid {
    flex-direction: column;
    gap: 1rem;
  }
  .vm-assignment-card {
    min-width: unset;
  }
}
.node-option-row {
  padding: .5rem;
  margin: .5rem;
  cursor: pointer;
}
.node-id {
  font-weight: 600;
  margin-bottom: 2px;
  margin-right: 1rem;
}
.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.chip {
  display: inline-block;
  background: #23243a;
  color: #b4befe;
  border-radius: 8px;
  padding: 2px 8px;
  margin-right: 4px;
  font-size: 0.85em;
}
</style> 