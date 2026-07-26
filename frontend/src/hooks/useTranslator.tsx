import { useEffect, useRef, useState } from "react";
import { translateStream } from "../api";
import type { TranslateResponse } from "../types";

export function useTranslator() {
  const [response, setResponse] = useState<TranslateResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [examplesLoading, setExamplesLoading] = useState(false);
  const controllerRef = useRef<AbortController | null>(null);
  const [initialQuery] = useState(
    () => new URLSearchParams(window.location.search).get("q") ?? ""
  );

  function handleTranslate(text: string) {
    controllerRef.current?.abort();
    setLoading(true);
    setExamplesLoading(false);
    setError(null);
    setResponse(null);

    const controller = translateStream(text, {
      onInit: (initResponse) => {
        setResponse(initResponse);
        setLoading(false);
        setExamplesLoading(true);
      },
      onExamples: (index, examples) => {
        setResponse((prev) => {
          if (!prev) {
            return prev;
          }
          const newResults = [...prev.results];
          newResults[index] = { ...newResults[index], examples };
          return { ...prev, results: newResults };
        });
      },
      onDone: () => {
        setExamplesLoading(false);
      },
      onError: (err) => {
        if (err.name === "AbortError") {
          return;
        }
        setError(err.message);
        setLoading(false);
        setExamplesLoading(false);
      },
    });

    controllerRef.current = controller;
  }

  useEffect(() => {
    return () => {
      controllerRef.current?.abort();
    };
  }, []);

  return {
    response,
    error,
    loading,
    examplesLoading,
    initialQuery,
    handleTranslate,
  };

}
