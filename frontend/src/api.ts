import type { TranslateResponse } from "./types";

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

