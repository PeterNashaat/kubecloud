import type { Node } from '../utils/userService';
import type { NormalizedNode } from '../types/normalizedNode';
import { computed } from 'vue';

export function useNormalizedNodes(nodes: () => Node[]) {
  return computed<NormalizedNode[]>(() => {
    return nodes().map((node: any) => {
      // Normalize price
      const price = node.price_usd ?? node.price ?? null;
      // Normalize GPU
      let gpu = 'none';
      if (node.num_gpu && node.num_gpu > 0) {
        gpu = 'present';
      }
      // Normalize location as a string
      let locationString = node.country;
      // Normalize CPU, RAM, Storage from total_resources
      const cpu = node.total_resources?.cru ?? 0;
      const ram = node.total_resources?.mru ? node.total_resources.mru / 1024 / 1024 / 1024 : 0; // in GB
      const storage = node.total_resources?.sru ? node.total_resources.sru / 1024 / 1024 / 1024 : 0; // in GB
      return {
        ...node,
        price,
        gpu,
        locationString,
        cpu,
        ram,
        storage,
      };
    });
  });
} 