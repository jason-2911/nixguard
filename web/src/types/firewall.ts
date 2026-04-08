// Firewall domain types — mirrors Go domain/firewall models

export interface FirewallRule {
  id: string;
  interface: string;
  direction: 'in' | 'out';
  action: 'pass' | 'block' | 'reject';
  protocol: string;
  source: Address;
  destination: Address;
  log: boolean;
  description: string;
  enabled: boolean;
  order: number;
  category: string;
  is_floating: boolean;
  interfaces?: string[];
  gateway?: string;
  state_type?: string;
  max_states?: number;
  tag?: string;
  tagged?: string;
  schedule?: Schedule;
  stats: RuleStats;
  created_at: string;
  updated_at: string;
}

export interface Address {
  type: 'any' | 'single' | 'network' | 'alias' | 'geoip';
  value: string;
  port: string;
  not: boolean;
}

export interface Schedule {
  name: string;
  start_time: string;
  end_time: string;
  weekdays: number[];
}

export interface RuleStats {
  evaluations: number;
  packets: number;
  bytes: number;
}

export interface FirewallAlias {
  id: string;
  name: string;
  type: 'host' | 'network' | 'port' | 'url' | 'url_table' | 'geoip' | 'nested';
  description: string;
  entries: string[];
  update_freq: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface NATRule {
  id: string;
  type: 'port_forward' | 'one_to_one' | 'outbound';
  interface: string;
  protocol: string;
  source: Address;
  destination: Address;
  redirect_target: string;
  redirect_port: string;
  description: string;
  enabled: boolean;
  nat_reflection: boolean;
  created_at: string;
}

export interface CapturedPacket {
  timestamp: string;
  interface: string;
  protocol: string;
  source: string;
  destination: string;
  length: number;
  verdict: string;
  summary: string;
  detail: string;
}

export interface PCAPExport {
  name: string;
  download_url: string;
  bytes: number;
  created_at: string;
}
