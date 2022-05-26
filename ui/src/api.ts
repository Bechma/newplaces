import EventEmitter, {
  DRAW_CANVAS,
  MESSAGE_SSE,
  SETUP_PALETTE,
} from "./eventEmitter";

export default class Api {
  eventSource!: EventSource;
  private readonly url: string;

  constructor(url: string) {
    this.url = url;
    this.init();
  }

  init() {
    this.eventSource = new EventSource(this.url + "events");
    this.eventSource.onopen = async (ev) => await this.onopen(ev);
    this.eventSource.onerror = (ev) => this.onerror(ev);
    this.eventSource.onmessage = (ev) => this.onmessage(ev);
  }

  async onopen(ev: Event) {
    console.log("Connected", ev);
    const response = await fetch(this.url + "canvas", {
      headers: {
        Accept: "application/octet-stream",
      },
    });
    const blob = await response.blob();
    const ab = await blob.arrayBuffer();
    const clamped = new Uint8ClampedArray(ab);
    EventEmitter.emit(DRAW_CANVAS, new ImageData(clamped, 2000, 2000));
  }

  onerror(ev: Event) {
    console.log("ERROR, closing", ev);
    this.eventSource.close();
    setTimeout(() => {
      console.log("Attempting to reconnect");
      this.init();
    }, 3000);
  }

  onmessage(ev: MessageEvent) {
    console.log("MESSAGE", ev);
    const { x, y, color } = JSON.parse(ev.data);
    EventEmitter.emit(MESSAGE_SSE, x, y, color);
  }

  sendPixel(x: number, y: number, color: number) {
    // TODO: Show the user whether the request has been ok
    fetch(this.url + "pixel", {
      method: "POST",
      body: JSON.stringify({ x, y, color }),
      headers: {
        "Content-Type": "application/json",
      },
    });
  }

  getPalette() {
    fetch(this.url + "palette", {
      headers: {
        Accept: "application/json",
      },
    }).then((res) => {
      res.json().then((res) => {
        EventEmitter.emit(SETUP_PALETTE, res);
      });
    });
  }
}
