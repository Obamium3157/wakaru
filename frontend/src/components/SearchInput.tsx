import { useState } from "react";
import styles from './SearchInput.module.css'

interface SearchInputProps {
  onSubmit: (text: string) => void;
  loading: boolean;
  initialValue?: string;
}

export function SearchInput({ onSubmit, loading, initialValue }: SearchInputProps) {
  const [value, setValue] = useState(initialValue ?? "");

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
      trimmed ? `/?q=${encodeURIComponent(trimmed)}` : window.location.pathname
    );
  }

  return (
    <form className={styles.searchForm} onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder="日本語の単語や文章を入力してください"
        value={value}
        onChange={handleChange}
        disabled={loading}
      />
    </form>
  )
}
