import { useLocation, useParams } from "react-router-dom";
import type { Example, Entry } from "../types";
import { useEffect, useState } from "react";
import { fetchWord } from "../api";

export function useWordInfo() {
  const { word } = useParams<{ word: string }>();
  const location = useLocation();
  const state = location.state as { examples?: Example[]; searchQuery?: string } | null;
  const examples = state?.examples ?? [];
  const searchQuery = state?.searchQuery ?? "";
  const posMajor = new URLSearchParams(location.search).get("pos") ?? undefined;

  const [entries, setEntries] = useState<Entry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!word) {
      setLoading(false);
      return;
    }
    let cancelled = false;
    fetchWord(word, posMajor)
      .then((data) => {
        if (!cancelled) {
          setEntries(data.entries);
        }
      })
      .catch((e) => {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "unknown error");
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => { cancelled = true; };
  }, [word, posMajor]);

  return {
    word,
    examples,
    searchQuery,
    entries,
    loading,
    error,
  }
}
