import "./style.css";
import CanvasState from "./canvasState";
import Api from "./api";
import EventEmitter, {
  DRAW_CANVAS,
  MESSAGE_SSE,
  SEND_PIXEL,
} from "./eventEmitter";

const canvasState = new CanvasState();
const sse = new Api("http://127.0.0.1:8000");

EventEmitter.subscribe(MESSAGE_SSE, (x, y, color) =>
  canvasState.paintPixel(x as number, y as number, color as number)
);
EventEmitter.subscribe(SEND_PIXEL, (x, y, color) => {
  sse.sendPixel(x as number, y as number, color as number);
});
EventEmitter.subscribe(DRAW_CANVAS, (array) =>
  canvasState.drawAllCanvas(array as ImageData)
);
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
}
