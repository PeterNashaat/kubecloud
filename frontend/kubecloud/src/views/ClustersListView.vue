<template>
  <div class="clusters-list-container">
    <div class="clusters-list-header">
      <h1 class="page-title">All Clusters</h1>
      <v-btn color="primary" @click="goToDeployCluster">
        <v-icon icon="mdi-plus" class="mr-1"></v-icon>
        New Cluster
      </v-btn>
    </div>
    <div class="clusters-list-toolbar">
      <v-text-field
        v-model="search"
        label="Search by name"
        prepend-inner-icon="mdi-magnify"
        clearable
        class="search-bar"
      />
      <v-select
        v-model="sortBy"
        :items="sortOptions"
        label="Sort by"
        class="filter-select"
      />
    </div>
    <v-divider class="mb-4" />
    <v-alert v-if="error" type="error" class="mb-4">{{ error }}</v-alert>
    <v-progress-linear v-if="isLoading" indeterminate color="primary" class="mb-4" />
    <div v-if="filteredClusters.length === 0 && !isLoading" class="empty-message">
      <v-icon icon="mdi-cloud-off-outline" size="48" class="mb-2" color="grey" />
      <div>No clusters found.</div>
    </div>
    <v-table v-else class="clusters-table">
      <thead>
        <tr>
          <th @click="setSort('name')" :class="sortBy === 'name' ? 'active-sort' : ''">Name <v-icon v-if="sortBy === 'name'" size="14">mdi-arrow-up-down</v-icon></th>
          <th @click="setSort('nodes')" :class="sortBy === 'nodes' ? 'active-sort' : ''">Nodes <v-icon v-if="sortBy === 'nodes'" size="14">mdi-arrow-up-down</v-icon></th>
          <th @click="setSort('createdAt')" :class="sortBy === 'createdAt' ? 'active-sort' : ''">Created <v-icon v-if="sortBy === 'createdAt'" size="14">mdi-arrow-up-down</v-icon></th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="cluster in filteredClusters" :key="cluster.id">
          <td class="cluster-name-cell">
            <span class="cluster-name">{{ cluster.project_name }}</span>
          </td>
          <td>{{ Array.isArray(cluster.cluster.nodes) ? cluster.cluster.nodes.length : (typeof cluster.cluster.nodes === 'number' ? cluster.cluster.nodes : 0) }}</td>
          <td>{{ formatDate(cluster.created_at) }}</td>
          <td>
            <v-btn icon size="small" class="mr-1" @click="viewCluster(cluster.project_name)"><v-icon icon="mdi-eye-outline" /></v-btn>
            <v-btn icon size="small" class="ml-1" color="error" @click="deleteCluster(cluster.project_name)"><v-icon icon="mdi-delete-outline" /></v-btn>
          </td>
        </tr>
      </tbody>
    </v-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useClusterStore } from '../stores/clusters';
import { useRouter } from 'vue-router';

const router = useRouter();
const clusterStore = useClusterStore();

const search = ref('');
const sortBy = ref('createdAt');

const sortOptions = [
  { value: 'name', title: 'Name' },
  { value: 'createdAt', title: 'Created' },
  { value: 'nodes', title: 'Nodes' },
];

const isLoading = computed(() => clusterStore.isLoading);
const error = computed(() => clusterStore.error);

onMounted(() => {
  clusterStore.fetchClusters();
});

const filteredClusters = computed(() => {
  let clusters = [...clusterStore.clusters];
  if (search.value) {
    clusters = clusters.filter(c => c.project_name.toLowerCase().includes(search.value.toLowerCase()));
  }
  // Sorting
  clusters.sort((a, b) => {
    if (sortBy.value === 'name') return a.project_name.localeCompare(b.project_name);
    if (sortBy.value === 'createdAt') return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
    // cost and nodes are not available from backend, skip for now
    return 0;
  });
  return clusters;
});

function setSort(field: string) {
  sortBy.value = field;
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr);
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function viewCluster(projectName: string) {
  router.push(`/clusters/${projectName}`);
}

async function deleteCluster(projectName: string) {
  if (confirm('Are you sure you want to delete this cluster? This action cannot be undone.')) {
    await clusterStore.deleteCluster(projectName);
    await clusterStore.fetchClusters();
  }
}

function goToDeployCluster() {
  router.push('/deploy');
}
</script>

<style scoped>
.clusters-list-container {
  max-width: 1200px;
  margin: 10rem auto;
  padding: 2rem 1rem;
}
.clusters-list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2rem;
}
.page-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text, #cfd2fa);
}
.clusters-list-toolbar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}
.search-bar {
  min-width: 220px;
  flex: 1 1 220px;
}
.filter-select {
  min-width: 160px;
}
.clusters-table {
  width: 100%;
  border-radius: 12px;
  overflow: hidden;
  background: var(--color-surface-1, #18192b);
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}
th, td {
  padding: 0.75rem 1rem;
  text-align: left;
}
th {
  background: var(--color-surface-2, #23243a);
  font-weight: 600;
  cursor: pointer;
  user-select: none;
}
th.active-sort {
  color: var(--color-primary, #6366f1);
}
tr {
  border-bottom: 1px solid var(--color-surface-2, #23243a);
}
tr:last-child {
  border-bottom: none;
}
.cluster-name-cell {
  font-weight: 600;
  color: var(--color-primary, #6366f1);
}
.empty-message {
  text-align: center;
  color: var(--color-text-muted, #7c7fa5);
  margin-top: 3rem;
}
@media (max-width: 900px) {
  .clusters-list-toolbar {
    flex-direction: column;
    gap: 0.5rem;
  }
  .clusters-list-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  .clusters-table {
    font-size: 0.95rem;
  }
}
.status-badge {
  display: inline-block;
  padding: 0.25em 0.75em;
  border-radius: 12px;
  font-size: 0.85em;
  font-weight: 600;
  text-transform: capitalize;
  background: var(--color-surface-2, #23243a);
  color: var(--color-text, #cfd2fa);
}
.status-badge.running {
  background: var(--color-success-subtle, #e6f9f0);
  color: var(--color-success, #22c55e);
}
.status-badge.stopped {
  background: var(--color-error-subtle, #fde8e8);
  color: var(--color-error, #ef4444);
}
.status-badge.deploying {
  background: var(--color-warning-subtle, #fff7e6);
  color: var(--color-warning, #f59e42);
}
</style> 