<template>
  <v-card>
    <v-card-title>
      <slot name="title" />
    </v-card-title>
    <v-card-text>
      <slot name="content" />
    </v-card-text>
    <v-card-actions v-if="$slots.actions">
      <slot name="actions" />
    </v-card-actions>
  </v-card>
</template>

<script lang="ts">
export default {};
</script> 