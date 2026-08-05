// Stub Wails runtime. Replaced by wailsjs/runtime during `wails generate module`.

export function EventsOn(eventName: string, callback: (...args: any[]) => void): void {
  console.log(`[Wails] EventsOn: ${eventName}`);
  // Store callback for possible emission during dev
  const handler = (e: CustomEvent) => callback(...(e.detail || []));
  window.addEventListener(`wails:${eventName}`, handler as EventListener);
}

export function EventsOff(eventName: string): void {
  console.log(`[Wails] EventsOff: ${eventName}`);
}

export function EventsOnce(eventName: string, callback: (...args: any[]) => void): void {
  console.log(`[Wails] EventsOnce: ${eventName}`);
}

export function EventsEmit(eventName: string, ...args: any[]): void {
  console.log(`[Wails] EventsEmit: ${eventName}`, args);
  window.dispatchEvent(new CustomEvent(`wails:${eventName}`, { detail: args }));
}

export function LogPrint(message: string): void {
  console.log(`[Wails] Log: ${message}`);
}
