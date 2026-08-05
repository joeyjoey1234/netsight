export function GetGreet(name: string): Promise<string> {
  console.log(`[Wails App] GetGreet(${name})`);
  return Promise.resolve(`Hello ${name} from NetSight!`);
}

export function StartScan(subnet: string, preset: string): Promise<string> {
  console.log(`[Wails App] StartScan(${subnet}, ${preset})`);
  return Promise.resolve('scan-001');
}

export function GetDevices(): Promise<any[]> {
  console.log('[Wails App] GetDevices');
  return Promise.resolve([]);
}

export function GetScanHistory(): Promise<any[]> {
  console.log('[Wails App] GetScanHistory');
  return Promise.resolve([]);
}

export function StartPacketCapture(iface: string, filter: string): Promise<string> {
  console.log(`[Wails App] StartPacketCapture(${iface}, ${filter})`);
  return Promise.resolve('capture-001');
}

export function RunPing(target: string, count: number): Promise<any> {
  console.log(`[Wails App] RunPing(${target}, ${count})`);
  return Promise.resolve(null);
}

export function RunTraceroute(target: string, mode: string): Promise<any[]> {
  console.log(`[Wails App] RunTraceroute(${target}, ${mode})`);
  return Promise.resolve([]);
}

export function StartServer(serverType: string, config: any): Promise<void> {
  console.log(`[Wails App] StartServer(${serverType})`, config);
  return Promise.resolve();
}

export function StopServer(serverType: string): Promise<void> {
  console.log(`[Wails App] StopServer(${serverType})`);
  return Promise.resolve();
}

export function ExportPDF(scanID: string): Promise<string> {
  console.log(`[Wails App] ExportPDF(${scanID})`);
  return Promise.resolve('');
}

export function ExportDrawIO(scanID: string): Promise<string> {
  console.log(`[Wails App] ExportDrawIO(${scanID})`);
  return Promise.resolve('');
}

export function GetNetworkInfo(): Promise<any> {
  console.log('[Wails App] GetNetworkInfo');
  return Promise.resolve(null);
}

export function CreateProject(name: string): Promise<any> {
  console.log(`[Wails App] CreateProject(${name})`);
  return Promise.resolve(null);
}

export function LoadProject(id: string): Promise<any> {
  console.log(`[Wails App] LoadProject(${id})`);
  return Promise.resolve(null);
}

export function WakeOnLAN(mac: string): Promise<void> {
  console.log(`[Wails App] WakeOnLAN(${mac})`);
  return Promise.resolve();
}

export function RunIPerf(target: string, serverMode: boolean, duration: number): Promise<any> {
  console.log(`[Wails App] RunIPerf(${target}, ${serverMode}, ${duration})`);
  return Promise.resolve(null);
}
