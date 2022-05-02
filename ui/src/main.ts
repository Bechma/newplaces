import "./style.css";
import CanvasState from "./canvasState";

const canvasState = new CanvasState();

setUpEventListeners();

function setUpEventListeners() {
  canvasState.camera.addEventListener("mousedown", (e) =>
    canvasState.initPan(e)
  );
  canvasState.camera.addEventListener("mouseup", () =>
    canvasState.setDragging(false)
  );
  canvasState.camera.addEventListener("mouseout", () =>
    canvasState.setDragging(false)
  );
  canvasState.camera.addEventListener("mousemove", (e) =>
    canvasState.panCamera(e)
  );
  canvasState.camera.addEventListener(
    "wheel",
    (e) => canvasState.toggleZoom(e),
    { passive: false }
  );
  canvasState.camera.addEventListener("click", () => canvasState.clickPixel());
}

function reddit(
  ctx: CanvasRenderingContext2D,
  location: string,
  dx: number,
  dy: number
) {
  const image = new Image();
  image.onload = () => ctx.drawImage(image, dx, dy);
  image.src = location;
}

function loadCanvas() {
  const ctx = canvasState.canvas.getContext("2d")!;
  reddit(ctx, "../img/reddit1.png", 0, 0);
  reddit(ctx, "../img/reddit2.png", 0, 1000);
  reddit(ctx, "../img/reddit3.png", 1000, 1000);
  reddit(ctx, "../img/reddit4.png", 1000, 0);
}

loadCanvas();
