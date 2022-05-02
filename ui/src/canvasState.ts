export default class CanvasState {
  x: number;
  y: number;
  scale: number;
  initialX: number;
  initialY: number;
  isDragging: boolean;
  panStartX: number;
  panStartY: number;
  clickedX: number;
  clickedY: number;
  middleX: number;
  middleY: number;
  camera: HTMLElement;
  position: HTMLElement;
  zoom: HTMLElement;
  selectedPixel: HTMLElement;
  canvas: HTMLCanvasElement;
  ctx: CanvasRenderingContext2D;

  constructor(
    x = 0,
    y = 0,
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
    this.scale = scale;
    this.initialX = initialX;
    this.initialY = initialY;
    this.isDragging = isDragging;
    this.panStartX = panStartX;
    this.panStartY = panStartY;
    this.clickedX = clickedX;
    this.clickedY = clickedY;
    this.middleX = middleX;
    this.middleY = middleY;
    this.camera = document.getElementById("camera")!;
    this.position = document.getElementById("position")!;
    this.zoom = document.getElementById("zoom")!;
    this.selectedPixel = document.getElementById("pixel")!;
    this.canvas = document.getElementById("canvas")! as HTMLCanvasElement;
    this.ctx = this.canvas.getContext("2d")!;
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
    this.clickedX = Math.floor((e.clientX - rect.left) / this.scale);
    this.clickedX = Math.min(Math.max(this.clickedX, 0), 1999);
    this.clickedY = Math.floor((e.clientY - rect.top) / this.scale);
    this.clickedY = Math.min(Math.max(this.clickedY, 0), 1999);

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
      if (this.scale * factor < 0.2) factor = 0.2 / this.scale;
    }
    // TODO: Fix camera jumps and zoom in to the border of the canvas, not outside
    this.scale *= factor;
    this.setX(middleX - (middleX - this.x) * factor);
    this.setY(middleY - (middleY - this.y) * factor);
    this.updateCamera();
  }

  clickPixel() {
    if (this.initialX == this.x && this.initialY == this.y) {
      this.ctx.fillStyle = "rgba(130,130,130,1)";
      this.ctx.fillRect(this.clickedX, this.clickedY, 1, 1);
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
