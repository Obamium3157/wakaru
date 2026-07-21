import type React from "react";
import { useState } from "react";

interface UseHandleSearchInputProps {
  onSubmit: (text: string) => void,
  initialValue?: string;
}

export function useHandleSearchInput({ onSubmit, initialValue }: UseHandleSearchInputProps) {
  const [value, setValue] = useState(initialValue ?? "")

  function handleSubmit(e: React.SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    const trimmed = value.trim();
    if (trimmed) {
      onSubmit(trimmed);
    }
  }

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const newValue = e.target.value;
    setValue(newValue);
    const trimmed = newValue.trim();
    window.history.replaceState(
      null,
      "",
      trimmed
        ? `/?q=${encodeURIComponent(trimmed)}`
        : window.location.pathname
    );
  }

  return {
    value,
    handleSubmit,
    handleChange,
  }
}
