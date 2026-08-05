// Stub Wails runtime. Replaced by wailsjs/runtime during `wails generate module`.

export function EventsOn(eventName: string, callback: (...args: any[]) => void): void {
  const handler = (e: CustomEvent) => callback(...(e.detail || []));
  window.addEventListener(`wails:${eventName}`, handler as EventListener);
}

export function EventsOff(eventName: string): void {
  // no-op stub
}

export function EventsOnce(eventName: string, callback: (...args: any[]) => void): void {
  // no-op stub
}

export function EventsEmit(eventName: string, ...args: any[]): void {
  window.dispatchEvent(new CustomEvent(`wails:${eventName}`, { detail: args }));
}

export function LogPrint(message: string): void {
  // no-op stub
}
