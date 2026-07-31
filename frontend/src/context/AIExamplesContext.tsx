import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { generateAIExamples } from "../api";
import type { AIExample } from "../types";

const STORAGE_KEY = "wakaru.ai-examples";

interface WordExamples {
  examples: AIExample[];
  loading: boolean;
  error: string | null;
}

interface AIExamplesContextValue {
  state: Record<string, WordExamples>;
  generate: (word: string) => void;
}

const emptyWordExamples: WordExamples = {
  examples: [],
  loading: false,
  error: null,
};

const AIExamplesContext = createContext<AIExamplesContextValue | null>(null);

function loadCache(): Record<string, WordExamples> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return {};
    }

    const parsed = JSON.parse(raw) as Record<string, AIExample[]>;
    const result: Record<string, WordExamples> = {};
    for (const [word, examples] of Object.entries(parsed)) {
      result[word] = { examples, loading: false, error: null };
    }

    return result;
  } catch {
    return {};
  }
}

export function AIExamplesProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<Record<string, WordExamples>>(loadCache);

  useEffect(() => {
    const cache: Record<string, AIExample[]> = {};
    for (const [word, entry] of Object.entries(state)) {
      if (entry.examples.length > 0) {
        cache[word] = entry.examples;
      }
    }
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(cache));
    } catch (e) {
      if (e instanceof Error) {
        console.log(e.message)
      }
    }
  }, [state]);

  const generate = useCallback((word: string) => {
    setState((prev) => ({
      ...prev,
      [word]: {
        examples: prev[word]?.examples ?? [],
        loading: true,
        error: null,
      },
    }));

    generateAIExamples(word)
      .then((data) => {
        setState((prev) => ({
          ...prev,
          [word]: { examples: data.examples, loading: false, error: null },
        }));
      })
      .catch((e) => {
        setState((prev) => ({
          ...prev,
          [word]: {
            examples: prev[word]?.examples ?? [],
            loading: false,
            error: e instanceof Error ? e.message : "unknown error",
          },
        }));
      });
  }, []);

  return (
    <AIExamplesContext.Provider value={{ state, generate }}>
      {children}
    </AIExamplesContext.Provider>
  );
}

export function useAIExamples(word: string) {
  const ctx = useContext(AIExamplesContext);
  if (!ctx) {
    throw new Error("useAIExamples must be used within AIExamplesProvider");
  }

  const wordState = ctx.state[word] ?? emptyWordExamples;

  return {
    examples: wordState.examples,
    loading: wordState.loading,
    error: wordState.error,
    generate: () => ctx.generate(word),
  };
}
