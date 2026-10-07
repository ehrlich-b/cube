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

export class CubeView {
  constructor(root, net, camera, stage) {
    this.root = root;
    this.net = net;
    this.camera = camera;
    this.pitch = -25;
    this.yaw = -33;
    this.reducedMotion = matchMedia("(prefers-reduced-motion: reduce)").matches;
    let drag;
    stage.addEventListener("pointerdown", event => {
      drag = { x: event.clientX, y: event.clientY, yaw: this.yaw, pitch: this.pitch };
      stage.setPointerCapture(event.pointerId);
    });
    stage.addEventListener("pointermove", event => {
      if (!drag) return;
      this.yaw = drag.yaw + (event.clientX - drag.x) * .5;
      this.pitch = Math.max(-80, Math.min(80, drag.pitch - (event.clientY - drag.y) * .5));
      this.orient();
    });
    stage.addEventListener("pointerup", () => { drag = null; });
    stage.addEventListener("pointercancel", () => { drag = null; });
    stage.addEventListener("keydown", event => {
      if (!event.key.startsWith("Arrow")) return;
      event.preventDefault();
      if (event.key === "ArrowLeft") this.yaw -= 15;
      if (event.key === "ArrowRight") this.yaw += 15;
      if (event.key === "ArrowUp") this.pitch = Math.max(-80, this.pitch - 15);
      if (event.key === "ArrowDown") this.pitch = Math.min(80, this.pitch + 15);
      this.orient();
    });
  }

  orient() { this.camera.style.transform = `rotateX(${this.pitch}deg) rotateY(${this.yaw}deg)`; }
  center() { this.pitch = -25; this.yaw = -33; this.orient(); }

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
    const layer = document.createElement("div");
    layer.className = "layer";
    for (const cubie of this.root.querySelectorAll(".cubie")) {
      if (selected(cubie.dataset.position.split(",").map(Number))) layer.append(cubie);
    }
    this.root.append(layer);
    const rotation = `rotate${["X", "Y", "Z"][axis]}(${angle}deg)`;
    const animation = layer.animate([{ transform: "none" }, { transform: rotation }], { duration, easing: "cubic-bezier(.3,.05,.25,1)", fill: "forwards" });
    await animation.finished;
    // The next render replaces the temporary layer with the engine's stickers.
  }
}
