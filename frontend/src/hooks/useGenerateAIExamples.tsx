import { useState } from "react";
import { generateAIExamples } from "../api";
import type { AIExample } from "../types";

interface GenerateAIExamplesProps {
  word?: string;
}

export const useGenerateAIExamples = ({ word }: GenerateAIExamplesProps) => {
  const [aiExamples, setAIExamples] = useState<AIExample[]>([]);
  const [aiLoading, setAILoading] = useState(false);
  const [aiError, setAIError] = useState<string | null>(null);

  const handleGenerateAIExamples = () => {
    if (!word) {
      return;
    }

    setAILoading(true);
    setAIError(null);
    generateAIExamples(word)
      .then((data) => setAIExamples(data.examples))
      .catch((e) => setAIError(e instanceof Error ? e.message : "unknown error"))
      .finally(() => setAILoading(false));
  };

  return {
    aiExamples,
    aiLoading,
    aiError,
    handleGenerateAIExamples,
  }
}
