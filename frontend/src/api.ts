import type { Entry, TranslateResponse } from "./types";

export async function translate(text: string): Promise<TranslateResponse> {
  const res = await fetch("/api/translate", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ text }),
  });

  if (!res.ok) {
    const message = await res.text();
    throw new Error(message || `request failed: ${res.status}`);
  }

  return res.json();
}

export async function fetchWord(text: string, posMajor?: string): Promise<{ entries: Entry[] }> {
  const params = new URLSearchParams();
  if (posMajor) {
    params.set("pos", posMajor);
  }

  const url = `/api/word/${encodeURIComponent(text)}${params.toString() ? "?" + params : ""}`;
  const res = await fetch(url);

  if (!res.ok) {
    const message = await res.text();
    throw new Error(message || `request failed: ${res.status}`);
  }

  return res.json();
}
