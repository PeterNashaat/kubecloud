<template>
  <div>
    <div class="section-header">
      <h3 class="section-title">
        <v-icon icon="mdi-check-circle" class="mr-2"></v-icon>
        Review Configuration
      </h3>
      <p class="section-subtitle">Review your cluster configuration before deployment</p>
    </div>
    <div class="review-grid">
      <div class="review-card">
        <h4 class="review-title">Node Assignment</h4>
        <div class="node-assignments">
          <div v-for="(vm, index) in allVMs" :key="index" class="node-assignment">
            <div class="assignment-icon">
              <v-icon :icon="vm.name.includes('Master') ? 'mdi-server' : 'mdi-desktop-tower'" :color="vm.name.includes('Master') ? 'var(--color-primary)' : 'var(--color-success)'" />
            </div>
            <div class="assignment-details">
              <div class="assignment-name">{{ vm.name }}</div>
              <div class="assignment-node">{{ getNodeInfo(vm.node !== null && vm.node !== undefined ? String(vm.node) : '') }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="step-actions">
      <v-btn variant="outlined" @click="$emit('prevStep')">
        <v-icon start icon="mdi-arrow-left"></v-icon>
        Back
      </v-btn>
      <v-btn color="success" :loading="deploying" @click="$emit('onDeployCluster')">
        <v-icon start icon="mdi-rocket-launch"></v-icon>
        Deploy Cluster
      </v-btn>
    </div>
  </div>
</template>
<script setup lang="ts">
import type { VM } from '../../composables/useDeployCluster';
import { defineProps, defineEmits } from 'vue';
const props = defineProps<{
  allVMs: VM[];
  getNodeInfo: (id: string) => string;
  onDeployCluster: () => void;
  prevStep: () => void;
  deploying: boolean;
}>();
const emit = defineEmits(['onDeployCluster', 'prevStep']);
</script>
<script lang="ts">
export default {
  name: 'Step3Review'
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
.review-grid {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
}
.review-card {
  background: var(--color-surface-1, #18192b);
  border-radius: 12px;
  padding: 1.5rem;
  margin-bottom: 2rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
  flex: 1 1 320px;
  min-width: 320px;
}
.review-title {
  font-size: 1.1rem;
  font-weight: 500;
  margin-bottom: 1rem;
}
.node-assignments {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}
.node-assignment {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--color-surface-2, #23243a);
}
.assignment-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary, #6366f1);
}
.assignment-details {
  display: flex;
  flex-direction: column;
}
.assignment-name {
  font-weight: 500;
  color: var(--color-text, #cfd2fa);
}
.assignment-node {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 0.98rem;
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
  .review-grid {
    flex-direction: column;
    gap: 1rem;
  }
  .review-card {
    min-width: unset;
  }
}
</style> 