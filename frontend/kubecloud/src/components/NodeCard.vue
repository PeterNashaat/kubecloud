<template>
  <div v-if="node" class="node-card" tabindex="0">
    <div class="node-header">
      <div class="node-title">{{ node.name }}</div>
      <div class="node-actions">
        <button class="node-action-btn" aria-label="Edit node" @click="$emit('edit')" title="Edit node">
          <v-icon icon="mdi-pencil" size="18" />
        </button>
        <button class="node-action-btn" aria-label="Remove node" @click="$emit('delete')" title="Remove node">
          <v-icon icon="mdi-delete" size="18" />
        </button>
      </div>
    </div>
    <div class="node-specs-grid">
      <div class="spec-item"><v-icon icon="mdi-cpu-64-bit" size="16" class="spec-icon" /> <span>vCPU</span> <span class="spec-value">{{ node.vcpu }}</span></div>
      <div class="spec-item"><v-icon icon="mdi-memory" size="16" class="spec-icon" /> <span>RAM</span> <span class="spec-value">{{ node.ram }} GB</span></div>
      <div class="spec-item"><v-icon icon="mdi-harddisk" size="16" class="spec-icon" /> <span>Disk</span> <span class="spec-value">{{ node.disk }} GB</span></div>
      <div class="spec-item"><v-icon icon="mdi-database" size="16" class="spec-icon" /> <span>Rootfs</span> <span class="spec-value">{{ node.rootfs }} GB</span></div>
    </div>
    <div class="node-divider"></div>
    <div class="node-ssh-chips">
      <template v-if="node.sshKeyIds && node.sshKeyIds.length">
        <span v-for="id in node.sshKeyIds" :key="id" class="ssh-chip">
          <v-icon icon="mdi-key-variant" size="14" class="ssh-chip-icon" />
          {{ getSshKeyName(id) }}
        </span>
      </template>
      <span v-else class="ssh-chip ssh-chip-none">No SSH Key</span>
    </div>
    <div class="node-badges-row">
      <span v-if="node.gpu" class="feature-badge" title="GPU enabled"><v-icon icon="mdi-nvidia" size="15" aria-label="GPU enabled" /> GPU</span>
      <span v-if="node.publicIp" class="feature-badge" title="Public IP enabled"><v-icon icon="mdi-earth" size="15" aria-label="Public IP enabled" /> Public IP</span>
      <span v-if="node.planetary" class="feature-badge" title="Planetary enabled"><v-icon icon="mdi-planet" size="15" aria-label="Planetary enabled" /> Planetary</span>
    </div>
  </div>
</template>
<script setup lang="ts">
import { defineProps } from 'vue';
import type { VM } from '../composables/useDeployCluster';
const props = defineProps<{
  node: VM;
  type: string;
  availableSshKeys: any[];
}>();
function getSshKeyName(id: number) {
  const key = props.availableSshKeys?.find((k: any) => k.id === id);
  return key ? key.name : 'Unknown';
}
</script>
<script lang="ts">
export default {
  name: 'NodeCard'
};
</script>
<style scoped>
.node-card {
  background: var(--color-bg-elevated);
  border-radius: 22px;
  padding: 2rem 1.5rem 1.5rem 1.5rem;
  margin-bottom: 1.5rem;
  box-shadow: var(--shadow-md);
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 1.1rem;
  outline: none;
}
.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.2rem;
}
.node-type-badge {
  background: var(--color-primary-subtle);
  color: var(--color-primary);
  font-size: 0.93rem;
  font-weight: 600;
  border-radius: 999px;
  padding: 0.13rem 1.1rem;
  letter-spacing: 0.02em;
  text-transform: uppercase;
  margin-right: 0.5rem;
}
.node-title {
  font-weight: 800;
  font-size: 1.22rem;
  margin-bottom: 0.2rem;
  letter-spacing: 0.01em;
}
.node-actions {
  display: flex;
  gap: 0.5rem;
}
.node-action-btn {
  background: none;
  border: none;
  color: var(--color-text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 7px;
  transition: background 0.15s, color 0.15s;
  outline: none;
}
.node-action-btn:focus, .node-action-btn:hover {
  background: var(--color-primary-subtle);
  color: var(--color-text);
}
.node-specs-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem 1.2rem;
  font-size: 1.01rem;
  color: var(--color-text-secondary);
  margin-bottom: 0.2rem;
}
.spec-item {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-weight: 500;
}
.spec-icon {
  opacity: 0.7;
}
.spec-value {
  margin-left: 0.25rem;
  font-weight: 600;
  color: var(--color-text-muted);
}
.node-divider {
  height: 1px;
  background: var(--color-border);
  margin: 0.7rem 0 0.2rem 0;
  border-radius: 1px;
}
.node-ssh-chips {
  margin-top: 0.7rem;
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.ssh-chip {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  background: var(--color-bg-hover);
  color: var(--color-text);
  border-radius: 999px;
  padding: 0.22rem 0.95rem;
  font-size: 1.01rem;
  font-weight: 600;
}
.ssh-chip-icon {
  opacity: 0.8;
}
.ssh-chip-none {
  background: var(--color-bg-hover);
  color: var(--color-text-muted);
  font-weight: 500;
}
.ssh-more {
  color: var(--color-text-muted);
  margin-left: 0.3rem;
  font-size: 0.93em;
}
.node-badges-row {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
}
.feature-badge {
  background: var(--color-primary-subtle);
  color: var(--color-primary);
  font-size: 0.93rem;
  border-radius: 999px;
  padding: 0.13rem 0.9rem;
  display: flex;
  align-items: center;
  gap: 0.2rem;
  font-weight: 600;
  text-transform: uppercase;
}
</style> 