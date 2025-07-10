import type { RawNode } from '../types/rawNode';
import type { NormalizedNode } from '../types/normalizedNode';

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
    // Add more UI fields as needed
  };
} 