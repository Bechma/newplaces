import "./style.css";
import CanvasState from "./canvasState";
import Api from "./api";
import EventEmitter, {
  COLOR_CHOOSER,
  DRAW_CANVAS,
  MESSAGE_SSE,
  SEND_PIXEL,
  SETUP_PALETTE,
} from "./eventEmitter";
import Palette from "./palette";

const canvasState = new CanvasState();
const sse = new Api("http://127.0.0.1:8080");
const palette = new Palette();

EventEmitter.subscribe(MESSAGE_SSE, (x, y, color) =>
  canvasState.paintPixel(x as number, y as number, color as number)
);
EventEmitter.subscribe(SEND_PIXEL, (x, y, color) => {
  sse.sendPixel(x as number, y as number, color as number);
});
EventEmitter.subscribe(DRAW_CANVAS, (array) =>
  canvasState.drawAllCanvas(array as ImageData)
);
EventEmitter.subscribe(SETUP_PALETTE, (array) =>
  palette.setUp(array as number[])
);
EventEmitter.subscribe(COLOR_CHOOSER, (a) => canvasState.setColor(a as number));
sse.getPalette();
setupCanvasEventListeners();

function setupCanvasEventListeners() {
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
  document
    .getElementById("deselect")!
    .addEventListener("click", () => canvasState.hidePx());
}
