export const colors = { W: "#fffcef", Y: "#f3d45d", R: "#df6352", O: "#efa254", B: "#6b9dcc", G: "#77ac78" };

// Rows and columns follow the engine's public face grids, viewed from outside.
export function stickerAddress(face, row, col, size = 3) {
  const h = (size - 1) / 2, a = col - h, b = h - row;
  return { F: [a, b, h], B: [-a, b, -h], L: [-h, b, a], R: [h, b, -a], U: [a, h, -b], D: [a, -h, b] }[face];
}

export function turnGeometry(token, size = 3) {
  const match = /^([1-7])?([URFDLB]w?|[MESxyz])('|2)?$/.exec(token);
  if (!match) throw new Error(`Invalid move ${token}`);
  const [, number, base, suffix] = match;
  const h = (size - 1) / 2;
  const [axis, sign, layer] = {
    R: [0, 1, h], L: [0, -1, -h], U: [1, 1, h], D: [1, -1, -h], F: [2, 1, h], B: [2, -1, -h],
    M: [0, -1, 0], E: [1, -1, 0], S: [2, 1, 0], x: [0, 1, null], y: [1, 1, null], z: [2, 1, null],
    Rw: [0, 1, h], Lw: [0, -1, -h], Uw: [1, 1, h], Dw: [1, -1, -h], Fw: [2, 1, h], Bw: [2, -1, -h]
  }[base];
  const wide = base.endsWith("w"), depth = Number(number || (wide ? 2 : 1));
  if (depth > size || ("MES".includes(base) && size % 2 === 0)) throw new Error(`Move ${token} is unavailable on ${size}×${size}`);
  const selectedLayer = layer === 0 || layer === null ? layer : layer - sign * (depth - 1);
  const angle = sign * (axis === 1 ? -90 : 90) * (suffix === "2" ? 2 : suffix === "'" ? -1 : 1);
  return { axis, angle, selected: position => layer === null || (wide ? position[axis] * sign >= h - depth + 1 : position[axis] === selectedLayer) };
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
    // Layout is sampled on orientation/resize, never on an animation tick.
    this.resizeObserver = new ResizeObserver(() => {
      this.projection = null;
      if (this.size > 3) this.drawSVG(Number(this.svg.dataset.turnAngle || 0));
    });
    this.resizeObserver.observe(stage);
  }

  cameraMatrix() { return new DOMMatrix().rotateAxisAngle(1, 0, 0, this.pitch).rotateAxisAngle(0, 1, 0, this.yaw); }
  orient() {
    this.projection = null;
    this.camera.style.transform = this.size > 3 ? "none" : `rotateX(${this.pitch}deg) rotateY(${this.yaw}deg)`;
    if (this.size > 3 && this.preview) this.turnCamera.style.transform = `rotateX(${this.pitch}deg) rotateY(${this.yaw}deg)`;
    if (this.size > 3) this.drawSVG(Number(this.svg.dataset.turnAngle || 0));
  }
  center() { this.cancelDrag(); this.pitch = -25; this.yaw = -33; this.orient(); }

  // Project the two face-plane tangents through the same CSS camera and
  // perspective as the stickers. This keeps swipe directions tied to the cube
  // even after orbiting to its back or underside.
  dragCandidates(sticker) {
    const face = sticker.dataset.face, index = Number(sticker.dataset.index);
    const position = stickerAddress(face, Math.floor(index / this.size), index % this.size, this.size);
    const normal = normals[face];
    const point = position.map((value, axis) => (value * this.unit + normal[axis] * (this.unit - 2) / 2) * (axis === 1 ? -1 : 1));
    const camera = this.size > 3 ? this.cameraMatrix() : new DOMMatrix(getComputedStyle(this.camera).transform);
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
      const h = (this.size - 1) / 2, sign = position[axis] < 0 ? -1 : 1;
      const depth = h - Math.abs(position[axis]) + 1;
      const faceToken = [["L", "R"], ["D", "U"], ["B", "F"]][axis][sign === 1 ? 1 : 0];
      // Resolve the picker once at pointerdown so preview and committed move
      // share exactly the same layer selection as buttons and keyboard turns.
      const token = this.selectMove ? this.selectMove(faceToken) : this.size === 3 ? layerMoves[axis][position[axis] + 1] :
        (depth === 1 ? "" : depth) + faceToken;
      const geometry = turnGeometry(token, this.size);
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
      drag.layer = this.makeLayer(drag.turn.geometry.selected, drag.turn.geometry.axis);
    }
    const { geometry, direction, pixelsPerDegree } = drag.turn;
    drag.degrees = Math.max(-180, Math.min(180, (dx * direction[0] + dy * direction[1]) / pixelsPerDegree));
    if (this.size > 3) this.drawSVG(drag.degrees * Math.sign(geometry.angle));
    else drag.layer.style.transform = rotation(geometry.axis, drag.degrees * Math.sign(geometry.angle));
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
      if (this.size > 3) {
        await this.animateSVG(drag.degrees * Math.sign(drag.turn.geometry.angle), target, this.reducedMotion ? 0 : 140);
        return;
      }
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

  svgNode(name, className) {
    const node = document.createElementNS("http://www.w3.org/2000/svg", name);
    node.setAttribute("class", className);
    return node;
  }

  solid(bounds, rotating = false) {
    return ["F", "B", "L", "R", "U", "D"].map(face => {
      const points = [[0, 0], [0, 1], [1, 1], [1, 0]].map(([row, col]) =>
        stickerAddress(face, row, col, 2).map((value, axis) =>
          (bounds[axis][0] + bounds[axis][1]) / 2 + value * (bounds[axis][1] - bounds[axis][0])));
      return { node: this.svgNode("polygon", "face cap"), points, rotating };
    });
  }

  // Six textured planes per slab replace hundreds of per-sticker surfaces.
  // Textures are rasterized once when a turn starts; CSS rotates the slab as
  // one group and the compositor reuses its pixels throughout the turn.
  texturedBlock(bounds) {
    const block = document.createElement("div");
    block.className = "turn-block";
    const bound = this.size * this.unit / 2 - 1;
    const scale = Math.min(devicePixelRatio || 1, 3);
    const angles = { F: "", B: "rotateY(180deg)", R: "rotateY(90deg)", L: "rotateY(-90deg)", U: "rotateX(90deg)", D: "rotateX(-90deg)" };
    for (const face of ["F", "B", "L", "R", "U", "D"]) {
      const normal = normals[face], axis = normal.findIndex(value => value);
      const center = bounds.map(([low, high]) => (low + high) / 2);
      center[axis] = bounds[axis][normal[axis] > 0 ? 1 : 0];
      const base = stickerAddress(face, 0, 0, this.size);
      const u = stickerAddress(face, 0, 1, this.size).map((value, i) => value - base[i]);
      const v = stickerAddress(face, 1, 0, this.size).map((value, i) => value - base[i]);
      const extent = direction => bounds.reduce((sum, [low, high], i) => sum + Math.abs(direction[i]) * (high - low), 0);
      const width = extent(u), height = extent(v);
      const canvas = document.createElement("canvas");
      canvas.className = "turn-surface";
      canvas.width = Math.ceil(width * scale); canvas.height = Math.ceil(height * scale);
      Object.assign(canvas.style, { width: `${width}px`, height: `${height}px`, marginLeft: `${-width / 2}px`, marginTop: `${-height / 2}px`,
        transform: `translate3d(${center[0]}px,${-center[1]}px,${center[2]}px) ${angles[face]}` });
      const context = canvas.getContext("2d");
      context.scale(canvas.width / width, canvas.height / height);
      context.fillStyle = "#111"; context.fillRect(0, 0, width, height);
      if (Math.abs(center[axis] - normal[axis] * bound) < .01) {
        const inset = 2 + Math.max(2, this.unit / 15);
        for (let index = 0; index < this.size * this.size; index++) {
          const position = stickerAddress(face, Math.floor(index / this.size), index % this.size, this.size);
          const delta = position.map((value, i) => value * this.unit - center[i]);
          const x = width / 2 + delta.reduce((sum, value, i) => sum + value * u[i], 0) - this.unit / 2;
          const y = height / 2 + delta.reduce((sum, value, i) => sum + value * v[i], 0) - this.unit / 2;
          context.fillStyle = "#24352e"; context.fillRect(x + 1, y + 1, this.unit - 2, this.unit - 2);
          context.strokeStyle = "#111"; context.lineWidth = .7; context.strokeRect(x + 1, y + 1, this.unit - 2, this.unit - 2);
          context.fillStyle = colors[this.stickers.get(`${face}:${index}`).dataset.color];
          context.fillRect(x + inset, y + inset, this.unit - 2 * inset, this.unit - 2 * inset);
          context.strokeStyle = "#24352e"; context.lineWidth = 1.6;
          context.strokeRect(x + inset, y + inset, this.unit - 2 * inset, this.unit - 2 * inset);
        }
      }
      block.append(canvas);
    }
    return block;
  }

  makeLayer(selected, axis) {
    if (this.size > 3) {
      const positions = [];
      for (const tile of this.svgTiles) {
        tile.rotating = selected(tile.position);
        if (tile.rotating) { tile.node.dataset.turning = "true"; positions.push(tile.position[axis]); }
      }
      const bound = this.size * this.unit / 2 - 1;
      const low = Math.max(-bound, (Math.min(...positions) - .5) * this.unit + 1);
      const high = Math.min(bound, (Math.max(...positions) + .5) * this.unit - 1);
      const block = (start, end, rotating) => {
        const bounds = [[-bound, bound], [-bound, bound], [-bound, bound]];
        bounds[axis] = [start, end];
        const slab = this.texturedBlock(bounds);
        if (rotating) this.preview = { axis, layer: slab };
        return slab;
      };
      const slabs = [block(low, high, true)];
      if (low > -bound) slabs.push(block(-bound, low - 2, false));
      if (high < bound) slabs.push(block(high + 2, bound, false));
      this.turnCamera.replaceChildren(...slabs);
      this.svg.style.display = "none";
      this.turnCamera.hidden = false;
      this.turnCamera.style.transform = `rotateX(${this.pitch}deg) rotateY(${this.yaw}deg)`;
      this.frameTimes = [];
      return this.preview;
    }
    const layer = document.createElement("div");
    layer.className = "layer";
    for (const cubie of this.root.querySelectorAll(".cubie")) {
      if (selected(cubie.dataset.position.split(",").map(Number))) layer.append(cubie);
    }
    this.root.append(layer);
    return layer;
  }

  projectionForView() {
    if (this.projection) return this.projection;
    const camera = this.cameraMatrix();
    const style = getComputedStyle(this.stage);
    const perspective = parseFloat(style.perspective);
    const origin = style.perspectiveOrigin.split(" ").map(parseFloat);
    const offset = [this.camera.offsetLeft - origin[0], this.camera.offsetTop - origin[1]];
    return this.projection = { camera, perspective, offset };
  }

  project(points, m) {
    const { perspective, offset } = this.projectionForView();
    return points.map(point => {
      const x = m.m11 * point[0] - m.m21 * point[1] + m.m31 * point[2];
      const y = m.m12 * point[0] - m.m22 * point[1] + m.m32 * point[2];
      const z = m.m13 * point[0] - m.m23 * point[1] + m.m33 * point[2];
      const scale = perspective / (perspective - z);
      return { x: x * scale + offset[0] * (scale - 1), y: y * scale + offset[1] * (scale - 1), z };
    });
  }

  projectItem(item, matrix) {
    const points = this.project(item.points, matrix);
    item.visible = points.reduce((sum, p, index) => {
      const next = points[(index + 1) % points.length];
      return sum + p.x * next.y - next.x * p.y;
    }, 0) > 0;
    item.depth = points.reduce((sum, p) => sum + p.z, 0) / 4;
    item.projected = points;
    if (item.sticker && item.visible) item.projectedInner = this.project(item.inner, matrix);
  }

  // SVG remains the idle hit surface. Animation changes one slab transform;
  // geometry and texture pixels stay fixed, with no per-frame DOM creation.
  drawSVG(angle = 0) {
    if (this.root.closest("[hidden]")) return;
    const measuring = this.root.hasAttribute("data-measure-frames");
    const started = measuring ? performance.now() : 0;
    if (this.preview) {
      this.preview.layer.style.transform = rotation(this.preview.axis, angle);
      this.svg.dataset.turnAngle = angle;
      if (measuring) {
        // Force style/layout as in the original SVG budget. Texture raster and
        // upload happen at turn setup; compositor paint is profiled separately.
        this.svg.getBBox();
        this.stage.getBoundingClientRect();
        this.frameTimes.push(performance.now() - started);
      }
      return;
    }
    const projectionChanged = !this.projection;
    const { camera } = this.projectionForView();
    this.turnCamera.hidden = true;
    this.svg.style.display = "";
    this.svg.dataset.turnAngle = angle;
    if (this.svgGeometry === this.projection) return;
    const text = points => points.map(p => `${p.x.toFixed(2)},${p.y.toFixed(2)}`).join(" ");
    const items = [...this.svgTiles, ...this.core];
    for (const item of items) {
      if (projectionChanged || !item.projected) this.projectItem(item, camera);
      item.node.style.display = item.visible ? "" : "none";
      if (!item.visible) continue;
      if (item.sticker) {
        item.plastic.setAttribute("points", text(item.projected));
        item.sticker.setAttribute("points", text(item.projectedInner));
      } else item.node.setAttribute("points", text(item.projected));
    }
    // Plastic is behind stickers; a large backing plane must not overpaint
    // far stickers merely because its center is nearer the camera.
    const caps = items.filter(item => !item.sticker).sort((a, b) => a.depth - b.depth);
    const tiles = items.filter(item => item.sticker).sort((a, b) => a.depth - b.depth);
    this.svg.append(...caps.map(item => item.node), ...tiles.map(item => item.node));
    this.svgGeometry = this.projection;
  }

  animateSVG(start, end, duration) {
    if (!duration) { this.drawSVG(end); return Promise.resolve(); }
    return new Promise(resolve => {
      const started = performance.now();
      const intervals = []; let last;
      const tick = now => {
        if (last !== undefined) intervals.push(now - last);
        last = now;
        const progress = Math.min(1, (now - started) / duration);
        const eased = progress * progress * (3 - 2 * progress);
        this.drawSVG(start + (end - start) * eased);
        if (progress < 1) requestAnimationFrame(tick);
        else {
          this.svg.dataset.frameTimes = JSON.stringify(this.frameTimes);
          this.svg.dataset.frameIntervals = JSON.stringify(intervals);
          resolve();
        }
      };
      requestAnimationFrame(tick);
    });
  }

  render(state) {
    const size = state.size || Math.sqrt(state.faces.F.length);
    if (size !== this.size) this.build(size);
    // Keep geometry and nodes between turns; only reattach a temporary layer
    // and update changed colors. A 7x7 has 294 stickers, independent of moves.
    for (const layer of this.root.querySelectorAll(".layer")) {
      this.root.append(...layer.children);
      layer.remove();
    }
    if (this.size > 3) {
      this.preview = null;
      this.turnCamera.replaceChildren();
      for (const tile of this.svgTiles) if (tile.rotating) { tile.rotating = false; delete tile.node.dataset.turning; }
    }
    for (const [face, stickers] of Object.entries(state.faces)) {
      stickers.forEach((color, index) => {
        const sticker = this.stickers.get(`${face}:${index}`);
        if (sticker.dataset.color !== color) {
          sticker.style.background = colors[color];
          if (this.size > 3) sticker.style.fill = colors[color];
          sticker.dataset.color = color;
          this.tiles.get(`${face}:${index}`).style.background = colors[color];
        }
      });
      this.netFaces.get(face).setAttribute("aria-label", `${face} face: ${stickers.join(" ")}`);
    }
    if (size > 3) this.drawSVG();
    this.root.dataset.solved = String(state.solved);
  }

  build(size) {
    this.size = size;
    this.unit = 186 / size;
    this.root.style.setProperty("--unit", `${this.unit}px`);
    this.root.style.setProperty("--half", `${(this.unit - 2) / 2}px`);
    this.net.style.setProperty("--size", size);
    this.root.dataset.size = size;
    this.stickers = new Map();
    this.tiles = new Map();
    this.netFaces = new Map();
    this.svgTiles = [];
    this.preview = null;
    this.projection = null;
    const cubies = new Map();
    const fragment = document.createDocumentFragment();
    const h = (size - 1) / 2;
    if (size <= 3) for (let x = -h; x <= h; x++) for (let y = -h; y <= h; y++) for (let z = -h; z <= h; z++) {
      const cubie = document.createElement("div");
      cubie.className = "cubie";
      cubie.dataset.position = `${x},${y},${z}`;
      cubie.style.transform = `translate3d(${x * this.unit}px,${-y * this.unit}px,${z * this.unit}px)`;
      for (const face of ["F", "B", "L", "R", "U", "D"]) {
        const surface = document.createElement("div");
        surface.className = `face ${face}`;
        cubie.append(surface);
      }
      cubies.set(`${x},${y},${z}`, cubie);
      fragment.append(cubie);
    }
    this.root.replaceChildren(fragment);
    if (size > 3) {
      this.svg = this.svgNode("svg", "svg-cube");
      this.svg.setAttribute("viewBox", "-180 -180 360 360");
      this.root.append(this.svg);
      this.core = this.solid([[-90, 90], [-90, 90], [-90, 90]]);
      this.turnCamera = document.createElement("div");
      this.turnCamera.className = "turn-camera";
      this.turnCamera.setAttribute("aria-hidden", "true");
      this.turnCamera.hidden = true;
      this.root.append(this.turnCamera);
    }
    this.camera.style.transform = size > 3 ? "none" : `rotateX(${this.pitch}deg) rotateY(${this.yaw}deg)`;
    this.net.replaceChildren();
    for (const face of ["F", "B", "L", "R", "U", "D"]) {
      const netFace = document.createElement("div");
      netFace.className = `net-face ${face}`;
      netFace.dataset.face = face;
      for (let index = 0; index < size * size; index++) {
        const position = stickerAddress(face, Math.floor(index / size), index % size, size);
        let sticker;
        if (size > 3) {
          const node = this.svgNode("g", "svg-tile");
          const plastic = this.svgNode("polygon", "plastic");
          sticker = this.svgNode("polygon", "sticker");
          node.append(plastic, sticker);
          const corners = inset => [[-inset, -inset], [-inset, inset], [inset, inset], [inset, -inset]].map(([row, col]) =>
            stickerAddress(face, Math.floor(index / size) + row, index % size + col, size).map((value, axis) =>
              value * this.unit + normals[face][axis] * (this.unit - 2) / 2));
          this.svgTiles.push({ node, plastic, sticker, position, rotating: false,
            points: corners(.5 - 1 / this.unit), inner: corners(.5 - (2 + Math.max(2, this.unit / 15)) / this.unit) });
          this.svg.append(node);
        } else {
          const surface = cubies.get(position.join(",")).querySelector(`.face.${face}`);
          sticker = document.createElement("div");
          sticker.className = "sticker";
          surface.append(sticker);
        }
        sticker.dataset.face = face;
        sticker.dataset.index = index;
        const tile = document.createElement("span");
        netFace.append(tile);
        this.stickers.set(`${face}:${index}`, sticker);
        this.tiles.set(`${face}:${index}`, tile);
      }
      this.net.append(netFace);
      this.netFaces.set(face, netFace);
    }
  }

  async animate(token, duration) {
    if (this.reducedMotion || this.root.closest("[hidden]")) return;
    const { axis, angle, selected } = turnGeometry(token, this.size);
    const layer = this.makeLayer(selected, axis);
    if (this.size > 3) { await this.animateSVG(0, angle, duration); return; }
    const animation = layer.animate([{ transform: "none" }, { transform: rotation(axis, angle) }], { duration, easing: "cubic-bezier(.3,.05,.25,1)", fill: "forwards" });
    await animation.finished;
    // The next render restores the layer nodes and updates their stickers.
  }
}
