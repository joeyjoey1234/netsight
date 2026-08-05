import { create } from 'zustand';
import type { ServerState } from '../types';

interface ServerStateMap {
  [type: string]: ServerState;
}

interface ServerStoreState {
  servers: ServerStateMap;
  updateServer: (server: ServerState) => void;
  clearAll: () => void;
}

export const useServerStore = create<ServerStoreState>((set) => ({
  servers: {},
  updateServer: (server) => set((state) => ({
    servers: { ...state.servers, [server.type]: server },
  })),
  clearAll: () => set({ servers: {} }),
}));
