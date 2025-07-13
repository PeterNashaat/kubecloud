import type { RawNode } from '../types/rawNode';
import type { NormalizedNode } from '../types/normalizedNode';
import type { RentedNode } from '../composables/useNodeManagement';

export function normalizeNode(node: RawNode): NormalizedNode {
  return {
    nodeId: node.nodeId,
    cpu: node.total_resources?.cru ?? 0,
    ram: node.total_resources?.mru ? Math.round(node.total_resources.mru / (1024 * 1024 * 1024)) : 0,
    storage: node.total_resources?.sru ? Math.round(node.total_resources.sru / (1024 * 1024 * 1024)) : 0,
    price_usd: typeof node.price_usd === 'number' ? node.price_usd : null,
    gpu: (node.num_gpu && node.num_gpu > 0) || (node.gpus && node.gpus.length > 0),
    locationString: node.country + (node.city ? ', ' + node.city : ''),
    country: node.country,
    city: node.city,
    status: node.status,
    healthy: node.healthy,
    rentable: node.rentable,
    rented: node.rented,
    dedicated: node.dedicated,
    certificationType: node.certificationType,
  };
}

export function getTotalCPU(node: RentedNode): number {
  if (!node) return 0;
  if (node.total_resources && typeof node.total_resources.cru === 'number') return node.total_resources.cru;
  if (node.resources && typeof node.resources.cpu === 'number') return node.resources.cpu;
  return 0;
}
export function getUsedCPU(node: RentedNode): number {
  if (!node) return 0;
  if (node.used_resources && typeof node.used_resources.cru === 'number') return node.used_resources.cru;
  return 0;
}
export function getAvailableCPU(node: RentedNode): number {
  if (!node) return 0;
  return Math.max(getTotalCPU(node) - getUsedCPU(node), 0);
}
export function getTotalRAM(node: RentedNode): number {
  if (!node) return 0;
  if (node.total_resources && typeof node.total_resources.mru === 'number') return Math.round(node.total_resources.mru / (1024 * 1024 * 1024));
  if (node.resources && typeof node.resources.memory === 'number') return node.resources.memory;
  return 0;
}
export function getUsedRAM(node: RentedNode): number {
  if (!node) return 0;
  if (node.used_resources && typeof node.used_resources.mru === 'number') return Math.round(node.used_resources.mru / (1024 * 1024 * 1024));
  return 0;
}
export function getAvailableRAM(node: RentedNode): number {
  if (!node) return 0;
  return Math.max(getTotalRAM(node) - getUsedRAM(node), 0);
}
export function getTotalStorage(node: RentedNode): number {
  if (!node) return 0;
  if (node.total_resources && typeof node.total_resources.sru === 'number') return Math.round(node.total_resources.sru / (1024 * 1024 * 1024));
  if (node.resources && typeof node.resources.storage === 'number') return node.resources.storage;
  return 0;
}
export function getUsedStorage(node: RentedNode): number {
  if (!node) return 0;
  if (node.used_resources && typeof node.used_resources.sru === 'number') return Math.round(node.used_resources.sru / (1024 * 1024 * 1024));
  return 0;
}
export function getAvailableStorage(node: RentedNode): number {
  if (!node) return 0;
  return Math.max(getTotalStorage(node) - getUsedStorage(node), 0);
} 