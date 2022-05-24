import { numberToRgba } from "./utils";
import EventEmitter, { COLOR_CHOOSER } from "./eventEmitter";

export default class Palette {
  palette: HTMLElement;

  constructor() {
    this.palette = document.getElementById("palette")!;
  }

  setUp(colors: number[]) {
    colors.forEach((a) => {
      const elem = document.createElement("div");
      elem.className = "picker";
      elem.style.background = numberToRgba(a);
      elem.onclick = () => EventEmitter.emit(COLOR_CHOOSER, a);
      this.palette.appendChild(elem);
    });
  }
}
