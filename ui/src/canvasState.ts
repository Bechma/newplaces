import EventEmitter, { SEND_PIXEL } from "./eventEmitter";
import { numberToRgba } from "./utils";

export default class CanvasState {
  x: number;
  y: number;
  color: number;
  scale: number;
  initialX: number;
  initialY: number;
  isDragging: boolean;
  panStartX: number;
  panStartY: number;
  clickedX: number;
  clickedY: number;
  originalClickedX: number;
  originalClickedY: number;
  middleX: number;
  middleY: number;
  camera: HTMLElement;
  position: HTMLElement;
  zoom: HTMLElement;
  selectedPixel: HTMLElement;
  canvas: HTMLCanvasElement;
  ctx: CanvasRenderingContext2D;
  px: HTMLElement;

  constructor(
    x = 0,
    y = 0,
    color = 0,
    scale = 1,
    initialX = 0,
    initialY = 0,
    isDragging = false,
    panStartX = 0,
    panStartY = 0,
    clickedX = 0,
    clickedY = 0,
    middleX = 0,
    middleY = 0
  ) {
    this.x = x;
    this.y = y;
    this.color = color;
    this.scale = scale;
    this.initialX = initialX;
    this.initialY = initialY;
    this.isDragging = isDragging;
    this.panStartX = panStartX;
    this.panStartY = panStartY;
    this.originalClickedX = this.clickedX = clickedX;
    this.originalClickedY = this.clickedY = clickedY;
    this.middleX = middleX;
    this.middleY = middleY;
    this.camera = document.getElementById("camera")!;
    this.position = document.getElementById("position")!;
    this.zoom = document.getElementById("zoom")!;
    this.selectedPixel = document.getElementById("pixel")!;
    this.canvas = document.getElementById("canvas")! as HTMLCanvasElement;
    this.ctx = this.canvas.getContext("2d")!;
    this.px = document.getElementById("pixel")!;
  }

  /**
   * Initialize the values required to move the canvas camera
   * @param e information about the mouse
   */
  initPan(e: MouseEvent) {
    this.panStartX = e.pageX;
    this.panStartY = e.pageY;
    this.initialX = this.x;
    this.initialY = this.y;
    this.isDragging = true;
  }

  /**
   * Move the camera based on the current position of the cursor, the initial
   * cursor position and where the position was at.
   * Also, it keeps track of the pixel that the cursor is pointing it and
   * the pixel of the canvas where the center of the screen is at.
   * @param e event
   */
  panCamera(e: MouseEvent) {
    const rect = this.canvas.getBoundingClientRect();
    this.originalClickedX = Math.floor((e.clientX - rect.left) / this.scale);
    this.clickedX = Math.min(Math.max(this.originalClickedX, 0), 1999);
    this.originalClickedY = Math.floor((e.clientY - rect.top) / this.scale);
    this.clickedY = Math.min(Math.max(this.originalClickedY, 0), 1999);

    this.middleX = Math.floor(
      (this.camera.offsetWidth / 2 - rect.left) / this.scale
    );
    this.middleY = Math.floor(
      (this.camera.offsetHeight / 4 - rect.top) / this.scale
    );
    if (this.isDragging) {
      this.setX(this.initialX + e.pageX - this.panStartX);
      this.setY(this.initialY + e.pageY - this.panStartY);
    }
    this.updateCamera();
  }

  /**
   * Change the zoom level of the camera
   * @param e wheel event
   */
  toggleZoom(e: WheelEvent) {
    const middleX = e.pageX - this.zoom.offsetWidth / 2,
      middleY = e.pageY - this.zoom.offsetHeight / 2;
    const scrollRatio = Math.abs(e.deltaY / 100) + 1;
    let factor;
    if (e.deltaY < 0) {
      factor = scrollRatio; // zoom in
      if (this.scale * factor > 35) factor = 35 / this.scale;
    } else {
      factor = 1 / scrollRatio; // zoom out is inverse of zoom in
      if (this.scale * factor < 0.5) factor = 0.5 / this.scale;
    }
    // TODO: Fix camera jumps and zoom in to the border of the canvas, not outside
    this.scale *= factor;
    this.setX(middleX - (middleX - this.x) * factor);
    this.setY(middleY - (middleY - this.y) * factor);
    this.updateCamera();
  }

  /**
   * If it's a valid pixel, we send the pixel along with the color to the api
   */
  clickPixel() {
    if (
      this.initialX == this.x &&
      this.initialY == this.y &&
      this.originalClickedY >= 0 &&
      this.originalClickedY < 2000 &&
      this.originalClickedX >= 0 &&
      this.originalClickedX < 2000 &&
      this.px.style.display != "none"
    ) {
      EventEmitter.emit(SEND_PIXEL, this.clickedX, this.clickedY, this.color);
    }
  }

  /**
   * Just set the dragging flag to the value provided
   * @param value
   */
  setDragging(value: boolean) {
    this.isDragging = value;
  }

  /**
   * Set the color of the pixel that we want to draw
   * @param color
   */
  setColor(color: number) {
    this.color = color;
    this.px.style.display = "block";
    this.px.style.backgroundColor = numberToRgba(color);
  }

  hidePx() {
    this.px.style.display = "none";
  }

  paintPixel(x: number, y: number, color: number) {
    this.ctx.fillStyle = numberToRgba(color);
    this.ctx.fillRect(x, y, 1, 1);
  }

  drawAllCanvas(ab: ImageData) {
    this.ctx.putImageData(ab, 0, 0);
  }

  /**
   * Update the position and zoom of the camera view
   * @private
   */
  private updateCamera() {
    this.position.style.transform = `translateX(${this.x}px) translateY(${this.y}px)`;
    this.zoom.style.transform = `scale(${this.scale})`;
    this.selectedPixel.style.left = `${this.clickedX}px`;
    this.selectedPixel.style.top = `${this.clickedY}px`;
  }

  /**
   * Move the x coordinate only if it's within the canvas limits(middle of the screen)
   * or the newX will put x into the canvas limits again.
   * @param newX new value for x
   * @private
   */
  private setX(newX: number) {
    if (
      (this.middleX >= 0 && this.middleX < 2000) ||
      (this.middleX < 0 && this.x > newX) ||
      (this.middleX >= 2000 && this.x <= newX)
    ) {
      this.x = newX;
    }
  }

  /**
   * Same functionality than `setX`
   * @param newY new value for y
   * @private
   */
  private setY(newY: number) {
    if (
      (this.middleY >= 0 && this.middleY < 2000) ||
      (this.middleY < 0 && this.y > newY) ||
      (this.middleY >= 2000 && this.y <= newY)
    ) {
      this.y = newY;
    }
  }
}
