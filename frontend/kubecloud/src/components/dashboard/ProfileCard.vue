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
          <label class="profile-label">Username</label>
          <v-text-field
            :model-value="user.username"
            variant="outlined"
            class="profile-field"
            color="accent"
            bg-color="transparent"
            hide-details="auto"
            readonly
          />
        </v-col>
        <v-col cols="12" md="6">
          <label class="profile-label">Email Address</label>
          <v-text-field
            :model-value="user.email"
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
      <v-row>
        <v-col cols="12" md="6">
          <label class="profile-label">Verified</label>
          <v-text-field
            :model-value="user.verified ? 'Yes' : 'No'"
            variant="outlined"
            class="profile-field"
            color="accent"
            bg-color="transparent"
            hide-details="auto"
            readonly
          />
        </v-col>
        <v-col cols="12" md="6">
          <label class="profile-label">Admin</label>
          <v-text-field
            :model-value="user.admin ? 'Yes' : 'No'"
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
        <v-col cols="12" md="6">
          <label class="profile-label">Credit Card Balance</label>
          <v-text-field
            :model-value="user.credit_card_balance"
            variant="outlined"
            class="profile-field"
            color="accent"
            bg-color="transparent"
            hide-details="auto"
            readonly
          />
        </v-col>
        <v-col cols="12" md="6">
          <label class="profile-label">Credited Balance</label>
          <v-text-field
            :model-value="user.credited_balance"
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
        <v-col cols="12" md="6">
          <label class="profile-label">Stripe Customer ID</label>
          <v-text-field
            :model-value="user.stripe_customer_id"
            variant="outlined"
            class="profile-field"
            color="accent"
            bg-color="transparent"
            hide-details="auto"
            readonly
          />
        </v-col>
        <v-col cols="12" md="6">
          <label class="profile-label">Last Updated</label>
          <v-text-field
            :model-value="formatDate(user.updated_at)"
            variant="outlined"
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
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'

const { user } = storeToRefs(useUserStore())

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleString()
}
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

.profile-label {
  display: block;
  margin-bottom: 0.25rem;
  color: --color-text;
  font-weight: 500;
  font-size: 0.95rem;
}
</style>
