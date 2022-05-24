import { numberToRgba } from "./utils";

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
      this.palette.appendChild(elem);
    });
  }
}
