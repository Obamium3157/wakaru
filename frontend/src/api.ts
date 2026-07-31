import type { AddBasicNoteRequest, AIExample, Entry, Example, TranslateResponse } from "./types";

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

type SSECallbacks = {
  onInit: (response: TranslateResponse) => void;
  onExamples: (index: number, examples: Example[]) => void;
  onDone: () => void;
  onError: (error: Error) => void;
}


export function translateStream(
  text: string,
  callbacks: SSECallbacks,
): AbortController {
  const controller = new AbortController();

  console.log('[translateStream] starting fetch for', text);
  fetch("/api/translate", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ text }),
    signal: controller.signal,
  })
    .then((res) => {
      console.log('[translateStream] response status:', res.status, 'for', text);
      checkResponse(res);
      return readSSEStream(res.body!.getReader(), callbacks);
    })
    .catch((err) => {
      console.log('[translateStream] caught:', err.name, err.message, 'for', text);
      callbacks.onError(err);
    });

  return controller;
}


function checkResponse(res: Response): void {
  if (!res.ok) {
    throw new Error(`request failed: ${res.status}`);
  }
}


async function readSSEStream(reader: ReadableStreamDefaultReader<Uint8Array>, callbacks: SSECallbacks): Promise<void> {
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }

    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split("\n\n");
    buffer = events.pop()!;

    for (const event of events) {
      const parsed = parseSSEEvent(event.split("\n"));
      if (parsed) {
        handleSSEEvent(parsed, callbacks);
      }
    }
  }
}


function parseSSEEvent(lines: string[]): { event: string; data: string } | null {
  let event = "";
  let data = "";

  for (const line of lines) {
    if (line.startsWith("event: ")) {
      event = line.slice(7);
    } else if (line.startsWith("data: ")) {
      data = line.slice(6);
    }
  }

  if (!event || !data) {
    return null;
  }

  return { event, data };
}


function handleSSEEvent(raw: { event: string; data: string }, callbacks: SSECallbacks): void {
  const parsed = JSON.parse(raw.data);
  switch (raw.event) {
    case "init":
      callbacks.onInit(parsed);
      break;
    case "examples":
      callbacks.onExamples(parsed.index, parsed.examples);
      break;
    case "done":
      callbacks.onDone();
      break;
  }
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

export async function addBasicNote(req: AddBasicNoteRequest): Promise<void> {
  const res = await fetch("/api/anki/note", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(req)
  });

  if (!res.ok) {
    const message = await res.text();
    throw new Error(message || `request failed: ${res.status}`);
  }
}

export async function generateAIExamples(word: string): Promise<{ examples: AIExample[] }> {
  const res = await fetch("/api/ai/examples", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ word })
  });

  if (!res.ok) {
    const message = await res.text();
    throw new Error(message || `request failed: ${res.status}`);
  }

  return res.json();
}
