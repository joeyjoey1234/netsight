export function EventsOn(eventName: string, callback: (...args: any[]) => void): void {
  console.log(`[Wails Runtime] EventsOn: ${eventName}`);
}

export function EventsOff(eventName: string): void {
  console.log(`[Wails Runtime] EventsOff: ${eventName}`);
}

export function EventsOnce(eventName: string, callback: (...args: any[]) => void): void {
  console.log(`[Wails Runtime] EventsOnce: ${eventName}`);
}

export function EventsEmit(eventName: string, ...args: any[]): void {
  console.log(`[Wails Runtime] EventsEmit: ${eventName}`, args);
}

export function LogPrint(message: string): void {
  console.log(`[Wails Runtime] Log: ${message}`);
}
