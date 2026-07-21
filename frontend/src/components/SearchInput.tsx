import styles from './SearchInput.module.css'
import { useHandleSearchInput } from "../hooks/useHandleSearchInput";

interface SearchInputProps {
  onSubmit: (text: string) => void;
  loading: boolean;
  initialValue?: string;
}

export function SearchInput({ onSubmit, loading, initialValue }: SearchInputProps) {
  const {
    value,
    handleSubmit,
    handleChange,
  } = useHandleSearchInput({ onSubmit, initialValue })

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
