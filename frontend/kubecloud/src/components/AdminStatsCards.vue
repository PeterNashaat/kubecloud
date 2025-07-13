<template>
  <div class="admin-section">
    <div class="section-header">
      <h2 class="dashboard-title">Admin Overview</h2>
      <p class="section-subtitle">Monitor platform health and key metrics</p>
    </div>
    <div class="stats-grid">
      <div v-for="stat in adminStats" :key="stat.label" class="stat-item">
        <v-icon :icon="stat.icon" size="24" :color="stat.color"></v-icon>
        <div class="stat-info">
          <div class="stat-number">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { defineProps } from 'vue'

interface AdminStat {
  icon: string;
  color: string;
  value: number;
  label: string;
}

const { adminStats } = defineProps<{ adminStats: AdminStat[] }>()
</script> 