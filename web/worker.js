import { loadEngine } from "./engine.js";

let engine;
// Register before loading finishes so the first request is never lost.
self.onmessage = async ({ data }) => {
  try {
    engine ??= loadEngine({ module: data.module, worker: !!data.request?.prepareFrames });
    const started = performance.now();
    self.postMessage({ type: "progress", phase: "loading" });
    const call = await engine;
    if (!data.request) return;
    const { request } = data;
    await call.prepare(request);
    self.postMessage({ type: "progress", phase: "solving" });
    const result = call(request);
    const computedMs = performance.now() - started;
    if (request.prepareFrames) {
      self.postMessage({ type: "progress", phase: "replay", completed: 0, total: result.moves.length });
      let cfen = request.cfen;
      // Send small verified batches instead of thousands of main-thread WASM
      // calls or one large structured clone. Cancellation terminates any phase.
      for (let index = 0; index < result.moves.length; index += 32) {
        const frames = call({ op: "sequence", cfen, moves: result.moves.slice(index, index + 32).join(" ") }).frames;
        cfen = frames.at(-1).cfen;
        self.postMessage({ type: "frames", frames, completed: Math.min(index + 32, result.moves.length), total: result.moves.length });
        await new Promise(resolve => setTimeout(resolve, 0));
      }
      if (cfen !== result.state.cfen) throw new Error("Sequence did not reach its verified checkpoint.");
    }
    self.postMessage({ ok: true, data: { ...result, computedMs } });
  } catch (error) {
    self.postMessage({ ok: false, error: error.message, code: error.code });
  }
};
