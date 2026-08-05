// Stub implementations matching the Go backend method signatures.
// Replaced by wailsjs/go/main/App during `wails generate module`.

export function GetGreet(name: string): Promise<string> {
  console.log(`[Wails] Greet(${name})`);
  return Promise.resolve(`Hello ${name} from NetSight!`);
}

export function StartScan(input: { subnet: string; preset: string }): Promise<string> {
  console.log(`[Wails] StartScan(${input.subnet}, ${input.preset})`);
  return Promise.resolve('scan-001');
}

export function StopScan(scanId: string): Promise<void> {
  console.log(`[Wails] StopScan(${scanId})`);
  return Promise.resolve();
}

export function GetDevices(): Promise<any[]> {
  console.log('[Wails] GetDevices');
  return Promise.resolve([]);
}

export function GetScanHistory(): Promise<any[]> {
  console.log('[Wails] GetScanHistory');
  return Promise.resolve([]);
}

export function StartPacketCapture(iface: string, filter: string): Promise<string> {
  console.log(`[Wails] StartPacketCapture(${iface}, ${filter})`);
  return Promise.resolve('capture-001');
}

export function StopPacketCapture(): Promise<void> {
  console.log('[Wails] StopPacketCapture');
  return Promise.resolve();
}

export function RunPing(input: { target: string; count: number }): Promise<any> {
  console.log(`[Wails] RunPing(${input.target}, ${input.count})`);
  return Promise.resolve({ target: input.target, latencyMs: 1.5, ttl: 64, bytes: 32, timedOut: false, sequence: 0 });
}

export function RunTraceroute(input: { target: string; mode: string }): Promise<any[]> {
  console.log(`[Wails] RunTraceroute(${input.target}, ${input.mode})`);
  return Promise.resolve([]);
}

export function RunNSLookup(query: string, types: string[]): Promise<any[]> {
  console.log(`[Wails] RunNSLookup(${query}, [${types}])`);
  return Promise.resolve([]);
}

export function StartServer(serverType: string, config: Record<string, any>): Promise<void> {
  console.log(`[Wails] StartServer(${serverType})`, config);
  return Promise.resolve();
}

export function StopServer(serverType: string): Promise<void> {
  console.log(`[Wails] StopServer(${serverType})`);
  return Promise.resolve();
}

export function ExportPDF(scanID: string): Promise<string> {
  console.log(`[Wails] ExportPDF(${scanID})`);
  return Promise.resolve('report.pdf');
}

export function ExportDrawIO(scanID: string): Promise<string> {
  console.log(`[Wails] ExportDrawIO(${scanID})`);
  return Promise.resolve('topology.drawio');
}

export function GetNetworkInfo(): Promise<any> {
  console.log('[Wails] GetNetworkInfo');
  return Promise.resolve({
    name: 'Ethernet',
    ips: ['192.168.1.100'],
    mac: 'aa:bb:cc:dd:ee:ff',
    gateway: '192.168.1.1',
    dns: ['8.8.8.8'],
    mtu: 1500,
  });
}

export function CreateProject(name: string): Promise<any> {
  console.log(`[Wails] CreateProject(${name})`);
  return Promise.resolve({ id: 'proj-001', name, created: new Date().toISOString() });
}

export function LoadProject(id: string): Promise<any> {
  console.log(`[Wails] LoadProject(${id})`);
  return Promise.resolve(null);
}

export function WakeOnLAN(mac: string): Promise<void> {
  console.log(`[Wails] WakeOnLAN(${mac})`);
  return Promise.resolve();
}

export function RunIPerf(input: { target: string; serverMode: boolean; duration: number }): Promise<any> {
  console.log(`[Wails] RunIPerf(${input.target}, server=${input.serverMode}, ${input.duration}s)`);
  return Promise.resolve({ bandwidthBps: 1000000000, transferBytes: 1250000000, interval: input.duration, jitterMs: 0.5 });
}
