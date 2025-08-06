<template>
  <v-card class="node-card mb-4" elevation="0">
    <div class="price-area px-4 pt-4 pb-2">
      <div class="d-flex align-center mb-1">
        <span style="color:#10B981; font-size:1.5rem; font-weight:700; letter-spacing:0.01em;">${{ node.price_usd ?? 'N/A' }}</span>
        <span class="text-caption ml-2" style="color:#a3a3a3; font-size:1.1rem; font-weight:500;">/month</span>
      </div>
      <div class="d-flex align-center">
        <span style="color:#10B981; font-size:1.1rem; font-weight:600;">${{ node.price_usd ? (Number(node.price_usd)/720).toFixed(2) : 'N/A' }}</span>
        <span class="text-caption ml-1" style="color:#a3a3a3; font-size:1.05rem; font-weight:500;">/hr</span>
      </div>
    </div>
    <div class="d-flex align-center justify-space-between px-4 pb-1">
      <span class="text-h6 font-weight-bold text-white">Node {{ node.nodeId }}</span>
      <v-chip v-if="node.gpu" color="#0ea5e9" variant="outlined" size="small" class="ml-2">GPU</v-chip>
    </div>
    <div v-if="node.country" class="d-flex align-center px-4 pb-1">
      <v-icon size="16" class="mr-1" color="#a3a3a3">mdi-map-marker</v-icon>
      <span class="text-body-2" style="color:#a3a3a3;">{{ node.country }}</span>
    </div>
    <v-card-text class="py-0 px-4">
      <div class="d-flex align-center mb-2 w-100">
        <v-icon size="18" color="#0ea5e9" class="mr-1">mdi-cpu-64-bit</v-icon>
        <span class="font-weight-medium" style="color:#a3a3a3; min-width:40px;">CPU:</span>
        <span class="text-white ml-1">{{ node.cpu }} vCPU</span>
      </div>
      <div class="d-flex align-center mb-2 w-100">
        <v-icon size="18" color="#10B981" class="mr-1">mdi-memory</v-icon>
        <span class="font-weight-medium" style="color:#a3a3a3; min-width:40px;">RAM:</span>
        <span class="text-white ml-1">{{ node.ram }} GB</span>
      </div>
      <div class="d-flex align-center mb-2 w-100">
        <v-icon size="18" color="#38bdf8" class="mr-1">mdi-harddisk</v-icon>
        <span class="font-weight-medium" style="color:#a3a3a3; min-width:40px;">Storage:</span>
        <span class="text-white ml-1">{{ node.storage }} GB</span>
      </div>
    </v-card-text>
    <v-card-actions class="pt-3 px-4 pb-4">
      <v-btn
        color="primary"
        variant="elevated"
        block
        class="font-weight-bold"
        style="border-radius:8px;"
        @click="$emit('reserve', node.nodeId)"
        aria-label="Reserve Node"
        :loading="loading"
        :disabled="disabled || loading"
      >
        Reserve Node
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import type { NormalizedNode } from '../types/normalizedNode';
import { defineProps, defineEmits } from 'vue';

const props = defineProps<{ node: NormalizedNode; loading?: boolean; disabled?: boolean }>();
const emit = defineEmits(['reserve', 'signin']);
</script>

<style scoped>
.node-card {
  border-radius: 14px;
  transition: box-shadow 0.2s, transform 0.2s;
}
.node-card:hover {
  transform: translateY(-2px) scale(1.01);
}
.price-area {
  background: rgba(16,185,129,0.07);
  border-radius: 12px;
  margin-bottom: 0.5rem;
  padding-top: 1rem;
  padding-bottom: 1rem;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}
</style>
