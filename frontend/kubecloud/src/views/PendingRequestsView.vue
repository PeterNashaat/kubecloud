<template>
  <div class="dashboard-container">
    <v-container fluid class="pa-0">
      <div class="dashboard-header mb-6">
        <h1 class="hero-title">Pending Requests</h1>
        <p class="section-subtitle">View and manage your pending transfer requests</p>
      </div>
      <div class="dashboard-content-wrapper">
        <div class="dashboard-layout">
          <!-- <div class="dashboard-sidebar">
            <DashboardSidebar :selected="'pending-requests'" @update:selected="handleSidebarSelect" />
          </div> -->
          <div class="dashboard-main">
            <div class="dashboard-cards">
              <PendingRequestsCard :pendingRequests="pendingRequests" />
            </div>
          </div>
        </div>
      </div>
    </v-container>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import PendingRequestsCard from '../components/dashboard/PendingRequestsCard.vue'
import type { PendingRequest } from '../utils/userService'

// Fake data for testing
const pendingRequests = ref<PendingRequest[]>([
  {
    created_at: new Date(Date.now() - 3600000 * 24 * 5).toISOString(),
    id: 1,
    tft_amount: 1000,
    transferred_tft_amount: 1000,
    transferred_usd_amount: 50,
    updated_at: new Date(Date.now() - 3600000 * 24 * 4).toISOString(),
    usd_amount: 50,
    user_id: 123
  },
  {
    created_at: new Date(Date.now() - 3600000 * 24 * 3).toISOString(),
    id: 2,
    tft_amount: 2500,
    transferred_tft_amount: 1250,
    transferred_usd_amount: 62.5,
    updated_at: new Date(Date.now() - 3600000 * 24 * 2).toISOString(),
    usd_amount: 125,
    user_id: 123
  },
  {
    created_at: new Date(Date.now() - 3600000 * 24 * 1).toISOString(),
    id: 3,
    tft_amount: 500,
    transferred_tft_amount: 0,
    transferred_usd_amount: 0,
    updated_at: new Date(Date.now() - 3600000 * 24 * 1).toISOString(),
    usd_amount: 25,
    user_id: 123
  },
  {
    created_at: new Date().toISOString(),
    id: 4,
    tft_amount: 3000,
    transferred_tft_amount: 0,
    transferred_usd_amount: 0,
    updated_at: new Date().toISOString(),
    usd_amount: 150,
    user_id: 123
  }
])
</script>

<style scoped>
.dashboard-container {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  background: var(--kubecloud-bg);
}

.hero-title {
  font-size: var(--font-size-4xl);
  font-weight: var(--font-weight-bold);
  margin-bottom: 1.5rem;
  line-height: 1.1;
  letter-spacing: -1px;
  color: var(--kubecloud-text);
}

.section-subtitle {
  font-size: var(--font-size-xl);
  color: var(--kubecloud-text-muted);
  line-height: 1.5;
  opacity: 0.92;
  margin-bottom: 0;
  font-weight: var(--font-weight-normal);
}

.dashboard-header {
  text-align: center;
  max-width: 900px;
  margin: 7rem auto 3rem auto;
  position: relative;
  z-index: 2;
  padding: 0 1rem;
}

.dashboard-content-wrapper {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 1rem;
  position: relative;
  z-index: 2;
  margin-top: 4rem;
}

.dashboard-layout {
  display: flex;
  min-height: 70vh;
  gap: 2.5rem;
  position: relative;
  z-index: 2;
  align-items: flex-start;
  margin-top: 0;
}

.dashboard-sidebar {
  flex: 0 0 280px;
  display: flex;
  flex-direction: column;
  height: fit-content;
  position: sticky;
  top: 0;
  align-self: flex-start;
  margin-top: 0;
}

.dashboard-main {
  flex: 1;
  min-width: 0;
}

.dashboard-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(380px, 1fr));
  gap: 2.2rem;
  width: 100%;
  align-items: stretch;
}
</style>
