// @ts-nocheck
import * as Backend from '../../wailsjs/go/wailsbridge/Bridge';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

export const backend = Backend;

export function onEvent(event: string, callback: (...args: any[]) => void): () => void {
  EventsOn(event, callback);
  return () => EventsOff(event);
}

export function startScan(subnet: string, preset: string): Promise<string> {
  return Backend.StartScan({ subnet, preset });
}

export function stopScan(scanId: string): Promise<void> {
  return Backend.StopScan(scanId);
}

export function startPacketCapture(iface: string, filter: string): Promise<string> {
  return Backend.StartPacketCapture(iface, filter);
}

export function stopPacketCapture(): Promise<void> {
  return Backend.StopPacketCapture();
}

export function runPing(target: string, count: number): Promise<any> {
  return Backend.RunPing({ target, count });
}

export function runTraceroute(target: string, mode: string): Promise<any[]> {
  return Backend.RunTraceroute({ target, mode });
}

export function runNSLookup(query: string, types: string[]): Promise<any[]> {
  return Backend.RunNSLookup(query, types);
}

export function wakeOnLAN(mac: string): Promise<void> {
  return Backend.WakeOnLAN(mac);
}

export function runIPerf(target: string, serverMode: boolean, duration: number): Promise<any> {
  return Backend.RunIPerf({ target, serverMode, duration });
}

export function startServer(type: string, config: Record<string, any>): Promise<void> {
  return Backend.StartServer(type, config);
}

export function stopServer(type: string): Promise<void> {
  return Backend.StopServer(type);
}

export function exportDrawIO(scanId: string): Promise<string> {
  return Backend.ExportDrawIO(scanId);
}

export function exportPDF(scanId: string): Promise<string> {
  return Backend.ExportPDF(scanId);
}

export function createProject(name: string): Promise<any> {
  return Backend.CreateProject(name);
}

export function loadProject(id: string): Promise<any> {
  return Backend.LoadProject(id);
}

export function getNetworkInfo(): Promise<any> {
  return Backend.GetNetworkInfo();
}

export function getAvailableSubnets(): Promise<string[]> {
  return Backend.GetAvailableSubnets();
}

export function getAllNetworkInfo(): Promise<any[]> {
  return Backend.GetAllNetworkInfo();
}

export function getDevices(): Promise<any[]> {
  return Backend.GetDevices();
}

export function getScanHistory(): Promise<any[]> {
  return Backend.GetScanHistory();
}

export function listProjects(): Promise<any[]> {
  return Backend.ListProjects();
}
