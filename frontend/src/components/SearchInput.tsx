import { useState } from "react";
import styles from './SearchInput.module.css'

interface SearchInputProps {
  onSubmit: (text: string) => void;
  loading: boolean;
}

export function SearchInput({ onSubmit, loading }: SearchInputProps) {
  const [value, setValue] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = value.trim();
    if (trimmed) {
      onSubmit(trimmed);
    }
  }

  return (
    <form className={styles.searchForm} onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder="日本語を入力..."
        value={value}
        onChange={(e) => setValue(e.target.value)}
        disabled={loading}
      />
      <button type="submit" disabled={loading || !value.trim()}>
        {loading ? "..." : "Translate"}
      </button>
    </form>
  )
}
