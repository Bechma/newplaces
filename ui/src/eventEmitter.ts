type func = (...args: unknown[]) => void;

export const DRAW_CANVAS = "DRAW_CANVAS",
  MESSAGE_SSE = "MESSAGE_SSE",
  SEND_PIXEL = "SEND_PIXEL";

export default class EventEmitter {
  static map: { [key: string]: func } = {};

  static emit(event: string, ...args: unknown[]): void {
    this.map[event] && this.map[event](...args);
  }

  static subscribe(event: string, handler: func): void {
    this.map[event] = handler;
  }
}
