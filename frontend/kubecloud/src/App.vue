<template>
  <ErrorBoundary>
    <v-app class="kubecloud-app">
      <!-- Loading overlay while checking authentication -->
      <div v-if="isInitializing" class="auth-loading-overlay">
        <v-progress-circular indeterminate color="primary" size="64"></v-progress-circular>
        <p class="loading-text">Initializing...</p>
      </div>

      <!-- Main app content -->
      <template v-else>
        <!-- Floating Cloud Animation - Site-wide -->
        <FloatingClouds />

        <!-- Shared Moving Background - Persists across route changes -->
        <UnifiedBackground :theme="currentTheme" />

        <!-- Navigation Bar -->
        <NavBar v-if="!isAuthPage" />

        <!-- Main Content Area -->
        <v-main class="app-main">
          <RouterView />
        </v-main>

        <!-- Footer -->
        <AppFooter v-if="!isAuthPage" />

        <!-- Global Notifications -->
        <NotificationToast />
      </template>
    </v-app>
  </ErrorBoundary>
</template>

<script lang="ts" setup>
import { RouterView, useRoute } from 'vue-router'
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import { useUserStore } from './stores/user'
import NavBar from './components/NavBar.vue'
import AppFooter from './components/AppFooter.vue'
import ErrorBoundary from './components/ErrorBoundary.vue'
import NotificationToast from './components/NotificationToast.vue'
import UnifiedBackground from './components/UnifiedBackground.vue'
import FloatingClouds from './components/FloatingClouds.vue'
import { useDeploymentEvents } from './composables/useDeploymentEvents'

const route = useRoute()
const userStore = useUserStore()
const isInitializing = ref(true)

// Determine if current page is an authentication page
const isAuthPage = computed(() => {
  const authRoutes = ['/sign-in', '/sign-up', '/register/verify']
  return authRoutes.includes(route.path)
})

// Dynamic theme based on current route
const currentTheme = computed(() => {
  const path = route.path

  // Theme mapping for different routes
  const themeMap: Record<string, 'default' | 'home' | 'features' | 'pricing' | 'use-cases' | 'docs' | 'nodes' | 'dashboard'> = {
    '/': 'home',
    '/features': 'features',
    '/pricing': 'pricing',
    '/usecases': 'use-cases',
    '/docs': 'docs',
    '/nodes': 'nodes',
    '/deploy': 'dashboard'
  }

  // Check for dashboard routes
  if (path.startsWith('/dashboard')) {
    return 'dashboard'
  }

  return themeMap[path] || 'default'
})

const { connect, disconnect } = useDeploymentEvents()

// Connect to deployment events only after user is logged in
watch(
  () => userStore.isLoggedIn,
  (loggedIn) => {
    if (loggedIn) {
      connect()
    } else {
      disconnect()
    }
  },
  { immediate: false }
)

// Initialize authentication state
onMounted(async () => {
  try {
    // Initialize auth state (check localStorage for tokens)
    await userStore.initializeAuth()
  } catch (error) {
    console.error('Failed to initialize authentication:', error)
  } finally {
    isInitializing.value = false
  }
})

onUnmounted(() => {
  disconnect()
})
</script>

<style scoped>
.kubecloud-app {
  min-height: 100vh;
  background: var(--color-bg);
  color: var(--color-text);
  font-family: 'Inter', sans-serif;
}

.app-main {
  position: relative;
  z-index: 1;
  min-height: calc(100vh - 72px); /* Account for navbar height */
}

.auth-loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--color-bg);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.loading-text {
  margin-top: 1rem;
  color: var(--color-text);
  font-size: 1.1rem;
  font-weight: 500;
}

/* Responsive adjustments */
@media (max-width: 960px) {
  .app-main {
    min-height: calc(100vh - 64px);
  }
}

@media (max-width: 600px) {
  .app-main {
    min-height: calc(100vh - 56px);
  }
}
</style>
