// Type declarations for Wails-generated bindings.
// These files are generated at build time by `wails generate module`.
// This file provides TypeScript with type information so the imports compile.

declare module '../../wailsjs/go/main/App' {
  export function StartScan(input: { subnet: string; preset: string }): Promise<string>;
  export function StopScan(scanId: string): Promise<void>;
  export function GetDevices(): Promise<any[]>;
  export function GetScanHistory(): Promise<any[]>;
  export function StartPacketCapture(iface: string, filter: string): Promise<string>;
  export function StopPacketCapture(): Promise<void>;
  export function RunPing(input: { target: string; count: number }): Promise<any>;
  export function RunTraceroute(input: { target: string; mode: string }): Promise<any[]>;
  export function RunNSLookup(query: string, types: string[]): Promise<any[]>;
  export function WakeOnLAN(mac: string): Promise<void>;
  export function RunIPerf(input: { target: string; serverMode: boolean; duration: number }): Promise<any>;
  export function StartServer(type: string, config: Record<string, any>): Promise<void>;
  export function StopServer(type: string): Promise<void>;
  export function ExportPDF(scanId: string): Promise<string>;
  export function ExportDrawIO(scanId: string): Promise<string>;
  export function CreateProject(name: string): Promise<any>;
  export function LoadProject(id: string): Promise<any>;
  export function ListProjects(): Promise<any[]>;
  export function GetNetworkInfo(): Promise<any>;
  export function GetAllNetworkInfo(): Promise<any[]>;
  export function GetAvailableSubnets(): Promise<string[]>;
  export function Greet(name: string): Promise<string>;
}

declare module '../../wailsjs/runtime' {
  export function EventsOn(event: string, callback: (...args: any[]) => void): void;
  export function EventsOff(event: string): void;
  export function EventsOnce(event: string, callback: (...args: any[]) => void): void;
  export function EventsEmit(event: string, ...args: any[]): void;
}
