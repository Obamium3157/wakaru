import { useState } from "react";
import { useSearchParams } from "react-router-dom";

interface UseHandleSearchInputProps {
  onSubmit: (text: string) => void,
  initialValue?: string;
}

export function useHandleSearchInput({ onSubmit }: UseHandleSearchInputProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const [value, setValue] = useState(searchParams.get("q") ?? "");

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
    setSearchParams(trimmed && trimmed
      ? { q: trimmed }
      : {},
      { replace: true }
    );
  }

  return {
    value,
    handleSubmit,
    handleChange,
  };
}
