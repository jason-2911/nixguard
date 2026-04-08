export interface NetworkRoute {
  id: string;
  destination: string;
  gateway: string;
  interface: string;
  metric: number;
  table: number;
  type: 'static' | 'dynamic' | 'policy';
  enabled: boolean;
  description: string;
  created_at?: string;
  updated_at?: string;
}

export interface GatewayStatus {
  state: string;
  latency_ms: number;
  packet_loss_percent: number;
  last_check?: string;
}

export interface Gateway {
  id: string;
  name: string;
  interface: string;
  address: string;
  protocol: string;
  monitor_ip: string;
  weight: number;
  priority: number;
  is_default: boolean;
  enabled: boolean;
  description: string;
  monitor_config: {
    interval: number;
    loss_threshold: number;
    latency_threshold: number;
    down_count: number;
    method: string;
  };
  status: GatewayStatus;
}

export interface GatewayGroupMember {
  gateway_id: string;
  tier: number;
  weight: number;
}

export interface GatewayGroup {
  id: string;
  name: string;
  members: GatewayGroupMember[];
  trigger: string;
  description: string;
}
