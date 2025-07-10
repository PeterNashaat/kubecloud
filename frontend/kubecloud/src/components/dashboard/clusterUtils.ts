import type { RentedNode } from '../../composables/useNodeManagement';

export function getClusterUsedResources(nodeId: number, editNodes: RentedNode[]): { vcpu: number, ram: number, storage: number } {
  // Sums up vcpu, ram, storage for all editNodes with this nodeId
  // editNodes may contain extended node objects with vcpu/ram/storage or cpu/memory/storage
  return (editNodes || []).filter((n: RentedNode) => n.nodeId === nodeId).reduce((acc: { vcpu: number, ram: number, storage: number }, n: RentedNode) => {
    acc.vcpu += ('vcpu' in n ? (n as any).vcpu : (n as any).cpu) || 0;
    acc.ram += ('ram' in n ? (n as any).ram : (n as any).memory) || 0;
    acc.storage += (n as any).storage || 0;
    return acc;
  }, { vcpu: 0, ram: 0, storage: 0 });
} 