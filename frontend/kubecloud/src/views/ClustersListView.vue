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
        v-model="statusFilter"
        :items="statusOptions"
        label="Status"
        clearable
        class="filter-select"
      />
      <v-select
        v-model="regionFilter"
        :items="regionOptions"
        label="Region"
        clearable
        class="filter-select"
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
          <th @click="setSort('status')" :class="sortBy === 'status' ? 'active-sort' : ''">Status <v-icon v-if="sortBy === 'status'" size="14">mdi-arrow-up-down</v-icon></th>
          <th @click="setSort('region')" :class="sortBy === 'region' ? 'active-sort' : ''">Region <v-icon v-if="sortBy === 'region'" size="14">mdi-arrow-up-down</v-icon></th>
          <th @click="setSort('nodes')" :class="sortBy === 'nodes' ? 'active-sort' : ''">Nodes <v-icon v-if="sortBy === 'nodes'" size="14">mdi-arrow-up-down</v-icon></th>
          <th @click="setSort('createdAt')" :class="sortBy === 'createdAt' ? 'active-sort' : ''">Created <v-icon v-if="sortBy === 'createdAt'" size="14">mdi-arrow-up-down</v-icon></th>
          <th @click="setSort('cost')" :class="sortBy === 'cost' ? 'active-sort' : ''">Est. Cost <v-icon v-if="sortBy === 'cost'" size="14">mdi-arrow-up-down</v-icon></th>
          <th>Tags</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="cluster in filteredClusters" :key="cluster.id">
          <td class="cluster-name-cell">
            <span class="cluster-name">{{ cluster.name }}</span>
          </td>
          <td>
            <v-chip :color="statusColor(cluster.status)" size="small" class="status-chip">{{ cluster.status }}</v-chip>
          </td>
          <td>{{ cluster.region }}</td>
          <td>{{ cluster.nodes }}</td>
          <td>{{ formatDate(cluster.createdAt) }}</td>
          <td>${{ cluster.cost.toLocaleString(undefined, { maximumFractionDigits: 2 }) }}</td>
          <td>
            <v-chip v-for="tag in cluster.tags" :key="tag" size="x-small" class="tag-chip">{{ tag }}</v-chip>
          </td>
          <td>
            <v-btn icon size="small" @click="viewCluster(cluster.id)"><v-icon icon="mdi-eye-outline" /></v-btn>
            <v-btn icon size="small" color="error" @click="deleteCluster(cluster.id)"><v-icon icon="mdi-delete-outline" /></v-btn>
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
const statusFilter = ref<string | null>(null);
const regionFilter = ref<string | null>(null);
const sortBy = ref('createdAt');

const statusOptions = ['running', 'stopped', 'starting', 'stopping', 'error'];
const sortOptions = [
  { value: 'name', title: 'Name' },
  { value: 'createdAt', title: 'Created' },
  { value: 'cost', title: 'Est. Cost' },
  { value: 'nodes', title: 'Nodes' },
  { value: 'region', title: 'Region' },
  { value: 'status', title: 'Status' },
];

const regionOptions = computed(() => {
  const regions = new Set(clusterStore.clusters.map(c => c.region));
  return Array.from(regions);
});

const isLoading = computed(() => clusterStore.isLoading);
const error = computed(() => clusterStore.error);

onMounted(() => {
  clusterStore.fetchClusters();
});

const filteredClusters = computed(() => {
  let clusters = [...clusterStore.clusters];
  if (search.value) {
    clusters = clusters.filter(c => c.name.toLowerCase().includes(search.value.toLowerCase()));
  }
  if (statusFilter.value) {
    clusters = clusters.filter(c => c.status === statusFilter.value);
  }
  if (regionFilter.value) {
    clusters = clusters.filter(c => c.region === regionFilter.value);
  }
  // Sorting
  clusters.sort((a, b) => {
    if (sortBy.value === 'name') return a.name.localeCompare(b.name);
    if (sortBy.value === 'createdAt') return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
    if (sortBy.value === 'cost') return b.cost - a.cost;
    if (sortBy.value === 'nodes') return b.nodes - a.nodes;
    if (sortBy.value === 'region') return a.region.localeCompare(b.region);
    if (sortBy.value === 'status') return a.status.localeCompare(b.status);
    return 0;
  });
  return clusters;
});

function setSort(field: string) {
  sortBy.value = field;
}

function statusColor(status: string) {
  switch (status) {
    case 'running': return 'success';
    case 'stopped': return 'error';
    case 'starting': return 'info';
    case 'stopping': return 'warning';
    case 'error': return 'error';
    default: return 'default';
  }
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr);
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function viewCluster(id: string | number) {
  router.push(`/clusters/${id}`);
}

function deleteCluster(id: string | number) {
  // TODO: Implement delete logic (confirmation, API call, update store)
  // For now, just log
  // eslint-disable-next-line no-console
  console.log('Delete cluster', id);
}

function goToDeployCluster() {
  router.push('/deploy');
}
</script>

<style scoped>
.clusters-list-container {
  max-width: 1200px;
  margin: 0 auto;
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
.status-chip {
  text-transform: capitalize;
}
.tag-chip {
  margin-right: 0.25rem;
  margin-bottom: 0.15rem;
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
</style> 