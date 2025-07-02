import { ref, computed } from 'vue';

export interface VM {
  name: string;
  vcpu: number;
  ram: number;
  node: number | null;
  rootfs: number;
  disk: number;
  gpu: boolean;
  sshKeyIds: number[];
  publicIp: boolean;
  planetary: boolean;
}
export interface Node { id: number; label: string; totalCPU: number; totalRAM: number; hasGPU: boolean; location: string; }
export interface SshKey { id: number; name: string; fingerprint: string; createdAt: string; }

export function useDeployCluster() {
  const masters = ref<VM[]>([]);
  const workers = ref<VM[]>([]);
  const availableNodes = ref<Node[]>([
    { id: 1234, label: 'Node #1234 (GPU)', totalCPU: 16, totalRAM: 64, hasGPU: true, location: 'us-east-1a' },
    { id: 5678, label: 'Node #5678', totalCPU: 8, totalRAM: 32, hasGPU: false, location: 'us-east-1b' },
    { id: 9012, label: 'Node #9012', totalCPU: 8, totalRAM: 32, hasGPU: false, location: 'us-east-1c' },
  ]);
  const availableSshKeys = ref<SshKey[]>([
    { id: 1, name: 'my-laptop-key', fingerprint: 'SHA256:abc123...', createdAt: '2024-01-15' },
    { id: 2, name: 'production-key', fingerprint: 'SHA256:def456...', createdAt: '2024-01-10' },
    { id: 3, name: 'team-shared-key', fingerprint: 'SHA256:ghi789...', createdAt: '2024-01-05' },
  ]);

  function addMaster() {
    if (masters.value.length < 3) {
      masters.value.push({
        name: `Master-${masters.value.length + 1}`,
        vcpu: 2,
        ram: 4,
        node: null,
        rootfs: 10,
        disk: 10,
        gpu: false,
        sshKeyIds: availableSshKeys.value.length ? [availableSshKeys.value[0].id] : [],
        publicIp: false,
        planetary: false,
      });
    }
  }
  function addWorker() {
    workers.value.push({
      name: `Worker-${workers.value.length + 1}`,
      vcpu: 2,
      ram: 4,
      node: null,
      rootfs: 10,
      disk: 10,
      gpu: false,
      sshKeyIds: availableSshKeys.value.length ? [availableSshKeys.value[0].id] : [],
      publicIp: false,
      planetary: false,
    });
  }
  function removeMaster(idx: number) {
    masters.value.splice(idx, 1);
  }
  function removeWorker(idx: number) {
    workers.value.splice(idx, 1);
  }

  return {
    masters, workers, availableNodes, availableSshKeys,
    addMaster, addWorker, removeMaster, removeWorker,
  };
} 