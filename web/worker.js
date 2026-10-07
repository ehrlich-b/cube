import { loadEngine } from "./engine.js";

const engine = loadEngine();
// Register before loading finishes so the first request is never lost.
self.onmessage = async ({ data }) => {
  try {
    const call = await engine;
    self.postMessage({ ok: true, data: call(data) });
  } catch (error) {
    self.postMessage({ ok: false, error: error.message });
  }
};
