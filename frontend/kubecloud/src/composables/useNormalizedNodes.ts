import type { RawNode } from '../types/rawNode';
import type { NormalizedNode } from '../types/normalizedNode';
import { computed } from 'vue';
import { normalizeNode } from '../utils/nodeNormalizer';

export function useNormalizedNodes(nodes: () => RawNode[]) {
  return computed<NormalizedNode[]>(() => nodes().map(normalizeNode));
} 