<template>
  <div class="dashboard-card">
    <div class="card-header">
      <div class="card-title-section">
        <div class="card-title-content">
          <h3 class="dashboard-card-title">Profile</h3>
          <p class="card-subtitle">Your account information</p>
        </div>
      </div>
    </div>
    <div v-if="user" class="profile-form">
      <v-row>
        <v-col cols="12" md="6">
          <v-text-field
            :model-value="profile.firstName"
            label="First Name"
            variant="outlined"
            class="profile-field"
            color="accent"
            bg-color="transparent"
            hide-details="auto"
            readonly
          />
        </v-col>
        <v-col cols="12" md="6">
          <v-text-field
            :model-value="profile.lastName"
            label="Last Name"
            variant="outlined"
            class="profile-field"
            color="accent"
            bg-color="transparent"
            hide-details="auto"
            readonly
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12">
          <v-text-field
            :model-value="profile.email"
            label="Email Address"
            variant="outlined"
            type="email"
            class="profile-field"
            color="accent"
            bg-color="transparent"
            hide-details="auto"
            readonly
          />
        </v-col>
      </v-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const user = computed(() => userStore.user)

const profile = ref({
  firstName: null as string | null,
  lastName: null as string | null,
  email: null as string | null
})

// Watch for user data and populate profile fields
watch(user, (newUser) => {
  if (newUser) {
    const [firstName, ...rest] = newUser.username.split(' ')
    profile.value.firstName = firstName
    profile.value.lastName = rest.join(' ')
    profile.value.email = newUser.email
  }
}, { immediate: true })
</script>

<style scoped>
.profile-form {
  width: 100%;
}

.profile-field {
  color: #CBD5E1;
  background: rgba(96, 165, 250, 0.08);
  border: 1px solid rgba(96, 165, 250, 0.12);
  border-radius: 0.75rem;
  margin-bottom: 1rem;
  transition: none;
  pointer-events: none; /* Prevents hover/focus */
}

.profile-field:focus-within,
.profile-field:hover {
  border-color: rgba(96, 165, 250, 0.12);
  background: rgba(96, 165, 250, 0.08);
}

.profile-field[readonly],
.profile-field :deep(input[readonly]) {
  background: rgba(96, 165, 250, 0.08) !important;
  border-color: rgba(96, 165, 250, 0.12) !important;
  cursor: default !important;
}

.profile-field :deep(.v-field) {
  background: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

.profile-field :deep(.v-field__input) {
  color: #fff !important;
  font-size: 1rem;
}

.profile-field :deep(.v-field__outline) {
  display: none !important;
}

.profile-field :deep(.v-label) {
  color: #60a5fa !important;
  font-weight: 500;
}

.profile-field :deep(.v-field--focused .v-label) {
  color: #60a5fa !important;
}

.profile-field :deep(.v-field--variant-outlined .v-field__outline__start) {
  border-color: transparent !important;
}

.profile-field :deep(.v-field--variant-outlined .v-field__outline__end) {
  border-color: transparent !important;
}

.profile-field :deep(.v-field--variant-outlined .v-field__outline__notch) {
  border-color: transparent !important;
}
</style>
