<template>
  <div class="card node-card">
    <div class="node-header">
      <h3 class="node-title">
        Node {{ node.nodeId }}
      </h3>
      <div class="node-price">${{ node.price ?? 'N/A' }}/month</div>
    </div>
    <div class="node-location d-flex justify-space-between" v-if="node.country">
      <div>
        <v-icon size="16" class="mr-1">mdi-map-marker</v-icon>
        {{ node.country }}
      </div>
      <div class="spec-item" v-if="node.gpu && node.gpu.toLowerCase() !== 'none'">
        <v-chip color="white" variant="outlined" size="small" class="mr-1">GPU</v-chip>
      </div>
    </div>
    <hr class="node-divider" />
    <div class="node-specs">
      <div class="spec-item">
        <v-icon size="18" class="mr-1" color="primary">mdi-cpu-64-bit</v-icon>
        <span class="spec-label">CPU:</span>
        <span>{{ Math.round(node.cpu) }} vCPU</span>
      </div>
      <div class="spec-item">
        <v-icon size="18" class="mr-1" color="success">mdi-memory</v-icon>
        <span class="spec-label">RAM:</span>
        <span>{{ Math.round(node.ram) }} GB</span>
      </div>
      <div class="spec-item">
        <v-icon size="18" class="mr-1" color="info">mdi-harddisk</v-icon>
        <span class="spec-label">Storage:</span>
        <span>{{ formatStorage(node.storage) }}</span>
      </div>
    </div>
    <v-btn 
      v-if="isAuthenticated"
      color="primary" 
      variant="elevated" 
      class="reserve-btn"
      @click="$emit('reserve', node.nodeId)"
      aria-label="Reserve Node"
    >
      Reserve Node
    </v-btn>
    <v-btn 
      v-else
      color="primary" 
      variant="outlined" 
      class="reserve-btn"
      @click="$emit('signin')"
      aria-label="Sign In to Reserve"
    >
      Sign In to Reserve
    </v-btn>
  </div>
</template>

<script setup lang="ts">
import type { NormalizedNode } from '../types/normalizedNode';
import { defineProps, defineEmits } from 'vue';

const props = defineProps<{ node: NormalizedNode; isAuthenticated: boolean }>();
const emit = defineEmits(['reserve', 'signin']);

function formatStorage(val: number) {
  if (val >= 1024) {
    return (val / 1024).toLocaleString(undefined, { maximumFractionDigits: 1, minimumFractionDigits: 1 }) + ' TB';
  }
  return Math.round(val).toLocaleString() + ' GB';
}
</script>

<style scoped>
.card.node-card {
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.06);
  padding: 1.5rem 1.25rem 1.25rem 1.25rem;
  margin: 0.5rem 0;
  transition: box-shadow 0.2s, transform 0.2s;
  min-height: 20rem;
  display: flex;
  flex-direction: column;
}
.card.node-card:hover {
  box-shadow: 0 6px 24px rgba(0,0,0,0.10);
  transform: translateY(-2px) scale(1.01);
}
.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}
.node-title {
  font-size: 1.15rem;
  font-weight: 600;
  margin: 0;
}
.node-price {
  font-size: 1.05rem;
  font-weight: 500;
  padding: 0.2rem 0.7rem;
}
.node-location {
  font-size: 0.98rem;
  color: #64748b;
}
.node-divider {
  border: none;
  margin: 0 0 1.1rem 0;
}
.node-specs {
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
}
.spec-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 1.01em;
  margin-bottom: 0.1rem;
}
.spec-label {
  font-size: 0.97em;
  color: #475569;
  min-width: 54px;
  font-weight: 500;
}
.reserve-btn {
  width: 100%;
  margin-top: auto;
  font-weight: 600;
  font-size: 1.05em;
  letter-spacing: 0.01em;
  border-radius: 8px;
  transition: background 0.18s, color 0.18s;
}
.reserve-btn[variant="elevated"] {
  background: #2563eb;
  color: #fff;
}
.reserve-btn[variant="elevated"]:hover {
  background: #1d4ed8;
}
.reserve-btn[variant="outlined"] {
  border: 1.5px solid #2563eb;
  color: #2563eb;
  background: #fff;
}
.reserve-btn[variant="outlined"]:hover {
  background: #e0e7ff;
  color: #1d4ed8;
}
</style>

// Add explicit default export for linter compatibility
export default {};