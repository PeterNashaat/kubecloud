import type { Node } from '../utils/userService';

export interface NormalizedNode extends Node {
  nodeId: number;
  cpu: number;
  ram: number;
  storage: number;
  price: number | null;
  gpu: string;
  locationString: string;
} 