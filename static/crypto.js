/**
 * Client-side AES-GCM encryption via Web Crypto API.
 * Key never leaves the browser (except as URL fragment).
 */

const te = new TextEncoder();
const td = new TextDecoder();

function bufToB64(buf) {
  const bytes = new Uint8Array(buf);
  let s = "";
  for (let i = 0; i < bytes.length; i++) s += String.fromCharCode(bytes[i]);
  return btoa(s);
}

function b64ToBuf(b64) {
  const s = atob(b64);
  const bytes = new Uint8Array(s.length);
  for (let i = 0; i < s.length; i++) bytes[i] = s.charCodeAt(i);
  return bytes.buffer;
}

function bytesToHex(buf) {
  return [...new Uint8Array(buf)].map((b) => b.toString(16).padStart(2, "0")).join("");
}

function hexToBytes(hex) {
  if (!/^[0-9a-fA-F]+$/.test(hex) || hex.length % 2 !== 0) {
    throw new Error("invalid key");
  }
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(hex.substr(i * 2, 2), 16);
  }
  return bytes.buffer;
}

async function generateKey() {
  const raw = crypto.getRandomValues(new Uint8Array(32));
  return {
    raw,
    hex: bytesToHex(raw),
    cryptoKey: await crypto.subtle.importKey(
      "raw",
      raw,
      { name: "AES-GCM" },
      false,
      ["encrypt", "decrypt"]
    ),
  };
}

async function importKeyFromHex(hex) {
  const raw = hexToBytes(hex);
  if (raw.byteLength !== 32) throw new Error("invalid key length");
  return crypto.subtle.importKey("raw", raw, { name: "AES-GCM" }, false, [
    "encrypt",
    "decrypt",
  ]);
}

/**
 * Encrypt plaintext. Returns base64(iv || ciphertext+tag).
 * IV is 12 random bytes prepended to the ciphertext.
 */
async function encryptText(plaintext, cryptoKey) {
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const cipher = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv },
    cryptoKey,
    te.encode(plaintext)
  );
  const combined = new Uint8Array(iv.byteLength + cipher.byteLength);
  combined.set(iv, 0);
  combined.set(new Uint8Array(cipher), iv.byteLength);
  return bufToB64(combined.buffer);
}

async function decryptText(ciphertextB64, cryptoKey) {
  const combined = new Uint8Array(b64ToBuf(ciphertextB64));
  if (combined.length < 13) throw new Error("ciphertext too short");
  const iv = combined.slice(0, 12);
  const data = combined.slice(12);
  const plain = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv },
    cryptoKey,
    data
  );
  return td.decode(plain);
}

window.QNotesCrypto = {
  generateKey,
  importKeyFromHex,
  encryptText,
  decryptText,
};
