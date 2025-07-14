<template>
  <div class="stats-grid">
    <div v-for="stat in stats" :key="stat.label" class="stat-item">
      <v-icon :icon="stat.icon" size="24" :color="stat.color || 'var(--color-primary)'"></v-icon>
      <div class="stat-info">
        <div class="stat-number">{{ stat.value }}</div>
        <div class="stat-label">{{ stat.label }}</div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { defineProps } from 'vue'

interface Stat {
  icon: string
  value: string | number
  label: string
  color?: string
}

const props = defineProps<{ stats: Stat[] }>()
const stats = props.stats
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: var(--space-4);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-number {
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-lg);
  color: var(--color-text);
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-primary);
}
</style>
