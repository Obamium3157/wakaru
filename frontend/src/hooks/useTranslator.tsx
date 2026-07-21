import { useState } from "react";
import { translate } from "../api";
import type { TranslateResponse } from "../types";

export function useTranslator() {
  const [response, setResponse] = useState<TranslateResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [initialQuery] = useState(
    () => new URLSearchParams(window.location.search).get("q") ?? ""
  );

  async function handleTranslate(text: string) {
    setLoading(true);
    setError(null);

    try {
      const data = await translate(text);
      setResponse(data);
    } catch (e) {
      setResponse(null);
      setError(e instanceof Error ? e.message : "unknown error");
    } finally {
      setLoading(false);
    }
  }

  return {
    response,
    error,
    loading,
    initialQuery,
    handleTranslate,
  }
}
