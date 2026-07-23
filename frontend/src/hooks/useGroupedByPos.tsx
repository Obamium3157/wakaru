import type { Translation } from "../types";

interface UseGroupedByPosProps {
  translations: Translation[];
}

export function useGroupedByPos({ translations }: UseGroupedByPosProps) {
  const groupedByPos = translations
    .reduce<Record<string, typeof translations>>((acc, t) => {
      const key = t.pos ?? "";

      if (!acc[key]) {
        acc[key] = [];
      }
      acc[key].push(t);

      return acc;
    }, {});

  return groupedByPos;
}
