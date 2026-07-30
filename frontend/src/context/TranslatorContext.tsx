import { createContext, useContext, useRef, useState, type ReactNode } from "react";
import type { TranslateResponse } from "../types";
import { translateStream } from "../api";

interface TranslationContextValue {
  response: TranslateResponse | null;
  error: string | null;
  loading: boolean;
  examplesLoading: boolean;
  handleTranslate: (text: string) => void;
}

const TranslationContext = createContext<TranslationContextValue | null>(null);

export function TranslationProvider({ children }: { children: ReactNode }) {
  const [response, setResponse] = useState<TranslateResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [examplesLoading, setExamplesLoading] = useState(false);
  const controllerRef = useRef<AbortController | null>(null);

  function handleTranslate(text: string) {
    controllerRef.current?.abort();
    setLoading(true);
    setExamplesLoading(true);
    setError(null);
    setResponse(null);

    controllerRef.current = translateStream(text, {
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
      onDone: () => setExamplesLoading(false),
      onError: (err) => {
        if (controllerRef.current?.signal.aborted) {
          return;
        }
        setError(err.message);
        setLoading(false);
        setExamplesLoading(false);
      }
    });
  }

  return (
    <TranslationContext.Provider
      value={{
        response,
        error,
        loading,
        examplesLoading,
        handleTranslate
      }}>
      {children}
    </TranslationContext.Provider>
  )
}

export function useTranslator() {
  const ctx = useContext(TranslationContext);
  if (!ctx) {
    throw new Error("useTranslation must be used within TranslationProvider");
  }
  return ctx;
}
