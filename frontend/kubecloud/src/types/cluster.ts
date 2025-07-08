// Types for the new backend Cluster and Node payload structure

export type NodeType = 'master' | 'worker' | 'leader';

export interface ClusterNode {
  name: string;
  type: NodeType;
  node_id: number;
  cpu: number;
  memory: number;      // MB
  root_size: number;    // MB
  disk_size: number;    // MB
  env_vars: Record<string, string>;
}

export interface Cluster {
  name: string;
  network: string;
  token: string;
  nodes: ClusterNode[];
} 