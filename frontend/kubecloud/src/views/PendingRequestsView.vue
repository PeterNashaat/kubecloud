<template>
  <div class="dashboard-container">
    <v-container fluid class="pa-0">
      <div class="dashboard-header mb-6">
        <h1 class="hero-title">Pending Requests</h1>
        <p class="section-subtitle">View your pending transfer requests</p>
      </div>
      <div class="dashboard-content-wrapper">
        <div class="dashboard-layout">
          <div class="dashboard-sidebar">
            <DashboardSidebar :selected="'pending-requests'" @update:selected="handleSidebarSelect" />
          </div>
          <div class="dashboard-main">
            <div class="dashboard-cards">
              <UserPendingRequestsCard />
            </div>
          </div>
        </div>
      </div>
    </v-container>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import UserPendingRequestsCard from '../components/dashboard/UserPendingRequestsCard.vue'

const selected = ref('overview')
function handleSidebarSelect(val: string) {
  selected.value = val
  // Save to localStorage for persistence
  localStorage.setItem('dashboard-section', val)
}

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
