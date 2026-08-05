import { useEffect } from 'react';

export function useWailsEvent(eventName: string, callback: (...args: any[]) => void) {
  useEffect(() => {
    const handler = (e: CustomEvent) => {
      if (e.detail) {
        callback(...e.detail);
      }
    };
    window.addEventListener(`wails:${eventName}`, handler as EventListener);
    return () => window.removeEventListener(`wails:${eventName}`, handler as EventListener);
  }, [eventName, callback]);
}

export async function callBackend(method: string, ...args: any[]): Promise<any> {
  console.log(`[Wails Call] ${method}`, args);
  return null;
}
