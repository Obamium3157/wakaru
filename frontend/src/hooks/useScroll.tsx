import { useState, useCallback } from "react";

export function useScroll<T>(array: T[]) {
  const [index, setIndex] = useState(0);
  const length = array.length;

  const scrollForwards = useCallback(() => {
    setIndex((prev) => (prev + 1) % length);
  }, [length]);

  const scrollBackwards = useCallback(() => {
    setIndex((prev) => (prev - 1 + length) % length);
  }, [length]);

  return { index, scrollForwards, scrollBackwards };
}
