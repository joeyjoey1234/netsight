import { create } from 'zustand';
import type { Scan, Device, Finding } from '../types';

interface ScanState {
  activeScanId: string | null;
  scanProgress: number;
  scanStatus: 'idle' | 'running' | 'completed' | 'failed';
  devices: Device[];
  scans: Scan[];
  findings: Finding[];
  setScanId: (id: string | null) => void;
  setProgress: (progress: number) => void;
  setStatus: (status: 'idle' | 'running' | 'completed' | 'failed') => void;
  addDevice: (device: Device) => void;
  setDevices: (devices: Device[]) => void;
  setScans: (scans: Scan[]) => void;
  addFinding: (finding: Finding) => void;
  reset: () => void;
}

export const useScanStore = create<ScanState>((set) => ({
  activeScanId: null,
  scanProgress: 0,
  scanStatus: 'idle',
  devices: [],
  scans: [],
  findings: [],
  setScanId: (id) => set({ activeScanId: id }),
  setProgress: (progress) => set({ scanProgress: progress }),
  setStatus: (status) => set({ scanStatus: status }),
  addDevice: (device) => set((state) => ({ devices: [...state.devices, device] })),
  setDevices: (devices) => set({ devices }),
  setScans: (scans) => set({ scans }),
  addFinding: (finding) => set((state) => ({ findings: [...state.findings, finding] })),
  reset: () => set({
    activeScanId: null,
    scanProgress: 0,
    scanStatus: 'idle',
    devices: [],
    findings: [],
  }),
}));
