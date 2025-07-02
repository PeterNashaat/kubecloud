<template>
  <div v-if="visible && localNode">
    <div class="modal-backdrop" @click="$emit('cancel')"></div>
    <div class="edit-node-modal">
      <h3>Edit Node</h3>
      <div class="modal-fields">
        <label>Name</label>
        <input type="text" v-model="localNode.name" />
        <div v-if="errors.name" class="field-error">{{ errors.name }}</div>
        <label>vCPU</label>
        <input type="number" v-model.number="localNode.vcpu" />
        <div v-if="errors.vcpu" class="field-error">{{ errors.vcpu }}</div>
        <label>RAM (GB)</label>
        <input type="number" v-model.number="localNode.ram" />
        <div v-if="errors.ram" class="field-error">{{ errors.ram }}</div>
        <label>Rootfs Size (GB)</label>
        <input type="number" v-model.number="localNode.rootfs" />
        <div v-if="errors.rootfs" class="field-error">{{ errors.rootfs }}</div>
        <label>Disk Size (GB)</label>
        <input type="number" v-model.number="localNode.disk" />
        <div v-if="errors.disk" class="field-error">{{ errors.disk }}</div>
        <label><input type="checkbox" v-model="localNode.gpu" /> GPU</label>
        <div class="ssh-key-section" style="margin-top: 1.5rem;">
          <label class="ssh-key-label">SSH Key</label>
          <v-chip-group
            :model-value="localNode.sshKeyIds"
            @update:model-value="ids => localNode.sshKeyIds = Array.isArray(ids) ? ids : [ids]"
            column
            style="max-width: 600px;"
            :multiple="false"
          >
            <v-chip
              v-for="key in availableSshKeys"
              :key="key.id"
              :value="key.id"
              class="ma-1"
              :class="{ 'selected-chip': localNode.sshKeyIds.includes(key.id) }"
            >
              {{ key.name }}
            </v-chip>
          </v-chip-group>
        </div>
        <div v-if="errors.ssh" class="field-error">{{ errors.ssh }}</div>
      </div>
      <div class="modal-actions">
        <button @click="onSave" :disabled="!valid">Save</button>
        <button @click="$emit('cancel')">Cancel</button>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { defineProps, defineEmits, watch, ref, computed } from 'vue';
import type { PropType } from 'vue';
import type { VM, SshKey } from '../composables/useDeployCluster';
const props = defineProps({
  node: { type: Object as PropType<VM>, required: true },
  visible: { type: Boolean, required: true },
  availableSshKeys: { type: Array as PropType<SshKey[]>, required: true }
});
const emit = defineEmits<{ (e: 'save', node: VM): void; (e: 'cancel'): void }>();
const localNode = ref<VM>({ ...props.node });
watch(() => props.node, (val) => { localNode.value = { ...val }; });
const errors = computed(() => {
  const node = localNode.value;
  const errs: Record<string, string> = {};
  if (!node.name || !node.name.trim()) errs.name = 'Name is required.';
  if (!node.vcpu || node.vcpu <= 0) errs.vcpu = 'vCPU must be a positive number.';
  if (!node.ram || node.ram <= 0) errs.ram = 'RAM must be a positive number.';
  if (!node.rootfs || node.rootfs <= 0) errs.rootfs = 'Rootfs size must be positive.';
  if (!node.disk || node.disk <= 0) errs.disk = 'Disk size must be positive.';
  if (!node.sshKeyIds || node.sshKeyIds.length === 0) errs.ssh = 'At least one SSH key must be selected.';
  return errs;
});
const valid = computed(() => Object.keys(errors.value).length === 0);
function onSave() {
  if (valid.value) emit('save', { ...localNode.value });
}
</script>
<script lang="ts">
export default {
  name: 'EditNodeModal'
};
</script>
<style scoped>
.modal-backdrop { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 1000; }
.edit-node-modal { position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%); background: #232946; color: #fff; border-radius: 16px; box-shadow: 0 8px 32px 0 rgba(0,0,0,0.25); padding: 2rem 2.5rem; z-index: 1001; min-width: 320px; max-width: 90vw; }
.edit-node-modal h3 { margin-top: 0; margin-bottom: 1.2rem; font-size: 1.3rem; font-weight: 700; }
.modal-fields label { display: block; margin-top: 1rem; font-weight: 500; }
.modal-fields input[type="number"], .modal-fields input[type="text"] { width: 100%; margin-top: 0.3rem; padding: 0.4em 0.7em; border-radius: 6px; border: 1px solid #4f8cff; background: #1a1f2b; color: #fff; }
.modal-actions { display: flex; gap: 1rem; margin-top: 2rem; justify-content: flex-end; }
.modal-actions button { background: #4f8cff; color: #fff; border: none; border-radius: 8px; padding: 0.5em 1.2em; font-size: 1em; cursor: pointer; transition: background 0.2s; }
.modal-actions button:last-child { background: #232946; border: 1px solid #4f8cff; color: #4f8cff; }
.field-error { color: #ff5252; font-size: 0.97em; margin-top: 0.2em; }
.ssh-key-section {
  margin-top: 1.5rem;
}
.ssh-key-label {
  font-weight: 600;
  margin-bottom: 0.5rem;
  display: block;
}
.v-chip.selected-chip {
  background: var(--color-primary, #6366f1);
  color: #fff;
}
</style> 