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
          :items="availableNodes"
          label="Select Reserved Node"
          v-model="vm.node"
          item-title="label"
          item-value="id"
          prepend-inner-icon="mdi-server-network"
          variant="outlined"
          :hint="vm.node !== null && vm.node !== undefined ? getNodeInfo(String(vm.node)) : 'Choose a node for this VM'"
          persistent-hint
          class="node-select"
        ></v-select>
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
  availableNodes: { id: string; label: string }[];
  getNodeInfo: (id: string) => string;
  onAssignNode: (vmIdx: number, nodeId: string) => void;
  isStep2Valid?: boolean;
}>(), {
  isStep2Valid: false
});
const emit = defineEmits(['nextStep', 'prevStep']);
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
</style> 