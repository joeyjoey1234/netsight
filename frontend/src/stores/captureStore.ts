import { create } from 'zustand';
import type { PacketSummary } from '../types';

interface CaptureState {
  isCapturing: boolean;
  packets: PacketSummary[];
  packetsPerSec: number;
  bytesPerSec: number;
  filter: string;
  setCapturing: (v: boolean) => void;
  addPacket: (packet: PacketSummary) => void;
  setStats: (pps: number, bps: number) => void;
  setFilter: (f: string) => void;
  clear: () => void;
}

export const useCaptureStore = create<CaptureState>((set) => ({
  isCapturing: false,
  packets: [],
  packetsPerSec: 0,
  bytesPerSec: 0,
  filter: '',
  setCapturing: (v) => set({ isCapturing: v }),
  addPacket: (packet) => set((state) => ({
    packets: [...state.packets.slice(-9999), packet],
  })),
  setStats: (pps, bps) => set({ packetsPerSec: pps, bytesPerSec: bps }),
  setFilter: (f) => set({ filter: f }),
  clear: () => set({ packets: [], packetsPerSec: 0, bytesPerSec: 0 }),
}));
