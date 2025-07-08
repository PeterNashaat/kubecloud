export interface NormalizedNode {
  nodeId: number;
  cpu: number; // vCPU
  ram: number; // GB
  storage: number; // GB
  price_usd: number | null;
  gpu: boolean;
  locationString: string;
  country: string;
  city: string;
  status: string;
  healthy: boolean;
  rentable: boolean;
  rented: boolean;
  dedicated: boolean;
  certificationType: string;
  // Add any other UI fields needed
} 