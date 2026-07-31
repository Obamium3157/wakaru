import { useAIExamples } from "../context/AIExamplesContext";

interface GenerateAIExamplesProps {
  word?: string;
}

export const useGenerateAIExamples = ({ word }: GenerateAIExamplesProps) => {
  const { examples, loading, error, generate } = useAIExamples(word ?? "");

  return {
    aiExamples: examples,
    aiLoading: loading,
    aiError: error,
    handleGenerateAIExamples: generate,
  }
}
