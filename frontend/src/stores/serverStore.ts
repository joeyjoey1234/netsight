import { create } from 'zustand';
import type { ServerState } from '../types';

interface ServerStateMap {
  [type: string]: ServerState;
}

interface ServerStoreState {
  servers: ServerStateMap;
  updateServer: (server: ServerState) => void;
  clearAll: () => void;
  hydrate: (servers: ServerState[] | Record<string, ServerState> | null | undefined) => void;
}

export const useServerStore = create<ServerStoreState>((set) => ({
  servers: {},
  updateServer: (server) => set((state) => ({
    servers: { ...state.servers, [server.type]: server },
  })),
  clearAll: () => set({ servers: {} }),
  hydrate: (servers) => set({ servers: Array.isArray(servers)
    ? Object.fromEntries(servers.filter(Boolean).map(server => [server.type, server]))
    : servers || {} }),
}));
