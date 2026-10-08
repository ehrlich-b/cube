// Go's nm does not recognize wasm. Inspect section/function sizes directly;
// build without -s when named function attribution is needed.
import { readFile } from "node:fs/promises";

const bytes = await readFile(process.argv[2] || "web/cube.wasm");
if (!bytes.subarray(0, 8).equals(Buffer.from([0, 97, 115, 109, 1, 0, 0, 0]))) throw new Error("Expected wasm version 1");
let position = 8;
const integer = () => {
  let value = 0, shift = 0, byte;
  do {
    if (position >= bytes.length || shift > 35) throw new Error("Truncated/oversized wasm integer");
    byte = bytes[position++];
    value += (byte & 127) * 2 ** shift;
    shift += 7;
  } while (byte & 128);
  return value;
};
const string = () => {
  const length = integer();
  const value = bytes.subarray(position, position + length).toString();
  position += length;
  return value;
};
const sections = [], functions = [], names = new Map();
let imports = 0;
while (position < bytes.length) {
  const id = bytes[position++], size = integer(), end = position + size;
  if (end > bytes.length) throw new Error("Truncated wasm section");
  let name = "";
  if (id === 0) {
    name = string();
    if (name === "name") while (position < end) {
      const subsection = bytes[position++], length = integer(), next = position + length;
      if (subsection === 1) {
        const count = integer();
        for (let i = 0; i < count; i++) names.set(integer(), string());
      }
      position = next;
    }
  } else if (id === 2) {
    const count = integer();
    for (let i = 0; i < count; i++) {
      string(); string();
      if (bytes[position++] !== 0) throw new Error("Expected Go's function-only imports");
      integer(); imports++;
    }
  } else if (id === 10) {
    const count = integer();
    for (let i = 0; i < count; i++) {
      const size = integer();
      functions.push({ index: imports + i, bytes: size });
      position += size;
    }
  }
  sections.push({ id, name, bytes: size });
  position = end;
}
for (const fn of functions) fn.name = names.get(fn.index) || `function[${fn.index}]`;
functions.sort((a, b) => b.bytes - a.bytes);
console.log(JSON.stringify({ raw: bytes.length, sections, topFunctions: functions.slice(0, 25) }, null, 2));
