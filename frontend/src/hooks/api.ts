// API layer — imports Wails backend stubs.
// During `wails build`, the real wailsjs bindings are generated and loaded.
// These stubs provide the same interface for development outside Wails.

import * as Backend from '../wails-stubs/main/App';
import { EventsOn, EventsOff } from '../wails-stubs/runtime';

export const backend = Backend;

export function onEvent(event: string, callback: (...args: any[]) => void): () => void {
  EventsOn(event, callback);
  return () => EventsOff(event);
}

// Wrappers that match Go struct parameter shapes

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

export function createProject(name: string): Promise<any> {
  return Backend.CreateProject(name);
}

export function loadProject(id: string): Promise<any> {
  return Backend.LoadProject ? Backend.LoadProject(id) : Promise.resolve(null);
}

export function getNetworkInfo(): Promise<any> {
  return Backend.GetNetworkInfo();
}

export function getAvailableSubnets(): Promise<string[]> {
  return Backend.GetAvailableSubnets ? Backend.GetAvailableSubnets() : Promise.resolve(['192.168.1.0/24']);
}

export function getAllNetworkInfo(): Promise<any[]> {
  return Backend.GetAllNetworkInfo ? Backend.GetAllNetworkInfo() :
    Backend.GetNetworkInfo().then((info: any) => Array.isArray(info) ? info : [info]);
}

export function getDevices(): Promise<any[]> {
  return Backend.GetDevices();
}

export function getScanHistory(): Promise<any[]> {
  return Backend.GetScanHistory();
}

export function listProjects(): Promise<any[]> {
  return Backend.ListProjects ? Backend.ListProjects() : Promise.resolve([]);
}
