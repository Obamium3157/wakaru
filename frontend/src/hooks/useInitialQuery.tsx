import { useState } from "react";

export const useInitialQuery = () => {
  const [initialQuery] = useState(() => new URLSearchParams(window.location.search).get("q") ?? "");

  return { initialQuery };
}
