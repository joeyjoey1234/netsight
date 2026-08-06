export interface Device {
  id: string;
  mac: string;
  ips: string[];
  vendor: string;
  hostname: string;
  os: string;
  role: 'switch' | 'router' | 'L3 switch' | 'server' | 'workstation' | 'unknown';
  vlans: number[];
  firstSeen: string;
  lastSeen: string;
  notes: string;
  model: string;
  links?: Link[];
}

export interface Port {
  deviceId: string;
  number: number;
  protocol: string;
  service: string;
  version: string;
  banner: string;
  state: 'open' | 'filtered' | 'closed';
}

export interface Link {
  sourceId: string;
  targetId: string;
  type: 'CDP' | 'LLDP' | 'ARP' | 'STP' | string;
  srcPort: string;
  dstPort: string;
  vlan: number;
}

export interface Scan {
  id: string;
  timestamp: string;
  subnet: string;
  duration: number;
  preset: 'quick' | 'short' | 'long' | 'manual';
  status: 'running' | 'completed' | 'failed';
  devicesFound: number;
  findings: Finding[];
}

export interface Finding {
  id: string;
  type: string;
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info';
  deviceId: string;
  title: string;
  description: string;
  recommendation: string;
  timestamp: string;
}

export interface Project {
  id: string;
  name: string;
  created: string;
  settings: ProjectSettings;
  devices?: Device[];
  scans?: Scan[];
  findings?: Finding[];
  links?: Link[];
}

export interface ProjectSettings {
  defaultSubnet: string;
  scanPorts: number[];
  excludeIps: string[];
}

export interface ServerState {
  type: string;
  port: number;
  status: 'stopped' | 'starting' | 'running' | 'error';
  interface: string;
  error?: string;
  startedAt?: string;
}

export interface InterfaceInfo {
  name: string;
  ips: string[];
  mac: string;
  gateway: string;
  dns: string[];
  publicIp: string;
  mtu: number;
  speed: number;
  duplex: string;
  crcErrors: number;
  collisions: number;
}

export interface PingResult {
  target: string;
  sequence: number;
  ttl: number;
  latencyMs: number;
  bytes: number;
  timedOut: boolean;
}

export interface Hop {
  number: number;
  ip: string;
  hostname: string;
  latencyMs: number;
  allIps: string[];
}

export interface PacketSummary {
  number: number;
  timestamp: string;
  srcMac: string;
  dstMac: string;
  srcIp: string;
  dstIp: string;
  protocol: string;
  srcPort?: number;
  dstPort?: number;
  length: number;
  info: string;
}

export interface IPerfResult {
  interval: number;
  transferBytes: number;
  bandwidthBps: number;
  jitterMs?: number;
  lostPackets?: number;
  totalPackets?: number;
}

export interface SubnetInfo {
  cidr: string;
  netmask: string;
  wildcard: string;
  network: string;
  broadcast: string;
  firstHost: string;
  lastHost: string;
  totalHosts: number;
  usableHosts: number;
  binary: string;
}
