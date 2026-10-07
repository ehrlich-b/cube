export const colors = { W: "#fffcef", Y: "#f3d45d", R: "#df6352", O: "#efa254", B: "#6b9dcc", G: "#77ac78" };

// Rows and columns follow the engine's public face grids, viewed from outside.
export function stickerAddress(face, row, col) {
  const a = col - 1, b = 1 - row;
  return { F: [a, b, 1], B: [-a, b, -1], L: [-1, b, a], R: [1, b, -a], U: [a, 1, -b], D: [a, -1, b] }[face];
}

export function turnGeometry(token) {
  const base = token.replace(/[2']/g, "");
  const [axis, sign, layer] = {
    R: [0, 1, 1], L: [0, -1, -1], U: [1, 1, 1], D: [1, -1, -1], F: [2, 1, 1], B: [2, -1, -1],
    M: [0, -1, 0], E: [1, -1, 0], S: [2, 1, 0], x: [0, 1, null], y: [1, 1, null], z: [2, 1, null],
    Rw: [0, 1, 1], Lw: [0, -1, -1], Uw: [1, 1, 1], Dw: [1, -1, -1], Fw: [2, 1, 1], Bw: [2, -1, -1]
  }[base];
  const angle = sign * (axis === 1 ? -90 : 90) * (token.endsWith("2") ? 2 : token.endsWith("'") ? -1 : 1);
  return { axis, angle, selected: position => layer === null || (base.endsWith("w") ? position[axis] * sign >= 0 : position[axis] === layer) };
}

const normals = { R: [1, 0, 0], L: [-1, 0, 0], U: [0, 1, 0], D: [0, -1, 0], F: [0, 0, 1], B: [0, 0, -1] };
const layerMoves = [["L", "M", "R"], ["D", "E", "U"], ["B", "S", "F"]];
const rotation = (axis, angle) => `rotate${["X", "Y", "Z"][axis]}(${angle}deg)`;

export class CubeView {
  constructor(root, net, camera, stage) {
    this.root = root;
    this.net = net;
    this.camera = camera;
    this.stage = stage;
    this.pitch = -25;
    this.yaw = -33;
    this.reducedMotion = matchMedia("(prefers-reduced-motion: reduce)").matches;
    stage.addEventListener("pointerdown", event => {
      if (event.button !== 0 || !event.isPrimary || this.drag) return;
      const sticker = event.target.closest(".sticker");
      if (sticker && !this.onTurnStart?.()) return;
      this.drag = { id: event.pointerId, x: event.clientX, y: event.clientY, yaw: this.yaw, pitch: this.pitch,
        candidates: sticker ? this.dragCandidates(sticker) : null };
      stage.dataset.dragging = sticker ? "turn" : "orbit";
      stage.focus({ preventScroll: true });
      event.preventDefault();
      stage.setPointerCapture(event.pointerId);
    });
    stage.addEventListener("pointermove", event => {
      if (event.pointerId !== this.drag?.id) return;
      this.moveDrag(event.clientX, event.clientY);
    });
    stage.addEventListener("pointerup", event => this.endDrag(event));
    stage.addEventListener("pointercancel", event => this.endDrag(event, true));
    stage.addEventListener("lostpointercapture", event => this.endDrag(event, true));
    window.addEventListener("blur", () => this.cancelDrag());
    stage.addEventListener("keydown", event => {
      if (event.key === "Escape" && this.drag) { event.preventDefault(); this.cancelDrag(); return; }
      if (!event.key.startsWith("Arrow")) return;
      event.preventDefault();
      if (this.drag?.candidates) return;
      if (event.key === "ArrowLeft") this.yaw -= 15;
      if (event.key === "ArrowRight") this.yaw += 15;
      if (event.key === "ArrowUp") this.pitch = Math.max(-80, this.pitch - 15);
      if (event.key === "ArrowDown") this.pitch = Math.min(80, this.pitch + 15);
      this.orient();
    });
  }

  orient() { this.camera.style.transform = `rotateX(${this.pitch}deg) rotateY(${this.yaw}deg)`; }
  center() { this.cancelDrag(); this.pitch = -25; this.yaw = -33; this.orient(); }

  // Project the two face-plane tangents through the same CSS camera and
  // perspective as the stickers. This keeps swipe directions tied to the cube
  // even after orbiting to its back or underside.
  dragCandidates(sticker) {
    const face = sticker.dataset.face, index = Number(sticker.dataset.index);
    const position = stickerAddress(face, Math.floor(index / 3), index % 3);
    const normal = normals[face];
    const point = position.map((value, axis) => (value * 62 + normal[axis] * 30) * (axis === 1 ? -1 : 1));
    const camera = new DOMMatrix(getComputedStyle(this.camera).transform);
    const style = getComputedStyle(this.stage);
    const perspective = parseFloat(style.perspective);
    const origin = style.perspectiveOrigin.split(" ").map(parseFloat);
    const project = matrix => {
      const p = camera.multiply(matrix).transformPoint(new DOMPoint(...point));
      const scale = perspective / (perspective - p.z);
      return [(p.x + this.camera.offsetLeft - origin[0]) * scale, (p.y + this.camera.offsetTop - origin[1]) * scale];
    };
    const start = project(new DOMMatrix());
    return [0, 1, 2].filter(axis => !normal[axis]).map(axis => {
      const token = layerMoves[axis][position[axis] + 1];
      const geometry = turnGeometry(token);
      const step = Math.sign(geometry.angle);
      const end = project(new DOMMatrix().rotate(axis === 0 ? step : 0, axis === 1 ? step : 0, axis === 2 ? step : 0));
      const tangent = end.map((value, i) => value - start[i]);
      const length = Math.hypot(...tangent);
      // A quarter turn takes roughly one sticker-to-sticker sweep across the
      // cube; normalize tiny foreshortened tangents to keep touch controllable.
      return { token, geometry, direction: tangent.map(value => value / length), pixelsPerDegree: Math.max(.65, Math.min(1.25, length)) };
    });
  }

  moveDrag(x, y) {
    const drag = this.drag;
    const dx = x - drag.x, dy = y - drag.y;
    if (!drag.candidates) {
      this.yaw = drag.yaw + dx * .5;
      this.pitch = Math.max(-80, Math.min(80, drag.pitch - dy * .5));
      this.orient();
      return;
    }
    if (!drag.turn) {
      if (Math.hypot(dx, dy) < 7) return;
      drag.turn = drag.candidates.reduce((best, candidate) =>
        Math.abs(dx * candidate.direction[0] + dy * candidate.direction[1]) > Math.abs(dx * best.direction[0] + dy * best.direction[1]) ? candidate : best);
      drag.layer = this.makeLayer(drag.turn.geometry.selected);
    }
    const { geometry, direction, pixelsPerDegree } = drag.turn;
    drag.degrees = Math.max(-180, Math.min(180, (dx * direction[0] + dy * direction[1]) / pixelsPerDegree));
    drag.layer.style.transform = rotation(geometry.axis, drag.degrees * Math.sign(geometry.angle));
  }

  endDrag(event, cancelled = false) {
    if (event.pointerId !== this.drag?.id) return;
    if (!cancelled) this.moveDrag(event.clientX, event.clientY);
    const drag = this.drag;
    this.drag = null;
    delete this.stage.dataset.dragging;
    if (this.stage.hasPointerCapture(drag.id)) this.stage.releasePointerCapture(drag.id);
    if (!drag.candidates) return;
    const quarters = cancelled || !drag.turn ? 0 : Math.round(Math.abs(drag.degrees) / 90);
    const token = quarters ? drag.turn.token + (quarters === 2 ? "2" : drag.degrees < 0 ? "'" : "") : null;
    const settle = async () => {
      if (!drag.layer) return;
      const target = quarters * 90 * Math.sign(drag.degrees) * Math.sign(drag.turn.geometry.angle);
      const animation = drag.layer.animate([
        { transform: drag.layer.style.transform }, { transform: rotation(drag.turn.geometry.axis, target) }
      ], { duration: this.reducedMotion ? 0 : 140, easing: "cubic-bezier(.2,.7,.2,1)", fill: "forwards" });
      await animation.finished;
    };
    this.onTurnEnd(token, settle);
  }

  cancelDrag() {
    if (this.drag) this.endDrag({ pointerId: this.drag.id }, true);
  }

  makeLayer(selected) {
    const layer = document.createElement("div");
    layer.className = "layer";
    for (const cubie of this.root.querySelectorAll(".cubie")) {
      if (selected(cubie.dataset.position.split(",").map(Number))) layer.append(cubie);
    }
    this.root.append(layer);
    return layer;
  }

  render(state) {
    const cubies = new Map();
    const fragment = document.createDocumentFragment();
    for (let x = -1; x <= 1; x++) for (let y = -1; y <= 1; y++) for (let z = -1; z <= 1; z++) {
      const cubie = document.createElement("div");
      cubie.className = "cubie";
      cubie.dataset.position = `${x},${y},${z}`;
      cubie.style.transform = `translate3d(${x * 62}px,${-y * 62}px,${z * 62}px)`;
      for (const face of ["F", "B", "L", "R", "U", "D"]) {
        const surface = document.createElement("div");
        surface.className = `face ${face}`;
        cubie.append(surface);
      }
      cubies.set(`${x},${y},${z}`, cubie);
      fragment.append(cubie);
    }
    this.root.replaceChildren(fragment);
    this.net.replaceChildren();
    for (const [face, stickers] of Object.entries(state.faces)) {
      const netFace = document.createElement("div");
      netFace.className = `net-face ${face}`;
      netFace.dataset.face = face;
      netFace.setAttribute("aria-label", `${face} face: ${stickers.join(" ")}`);
      stickers.forEach((color, index) => {
        const position = stickerAddress(face, Math.floor(index / 3), index % 3);
        const surface = cubies.get(position.join(",")).querySelector(`.face.${face}`);
        const sticker = document.createElement("div");
        sticker.className = "sticker";
        sticker.style.background = colors[color];
        sticker.dataset.color = color;
        sticker.dataset.face = face;
        sticker.dataset.index = index;
        surface.append(sticker);
        const tile = document.createElement("span");
        tile.style.background = colors[color];
        netFace.append(tile);
      });
      this.net.append(netFace);
    }
    this.root.dataset.solved = String(state.solved);
  }

  async animate(token, duration) {
    if (this.reducedMotion || this.root.closest("[hidden]")) return;
    const { axis, angle, selected } = turnGeometry(token);
    const layer = this.makeLayer(selected);
    const animation = layer.animate([{ transform: "none" }, { transform: rotation(axis, angle) }], { duration, easing: "cubic-bezier(.3,.05,.25,1)", fill: "forwards" });
    await animation.finished;
    // The next render replaces the temporary layer with the engine's stickers.
  }
}
