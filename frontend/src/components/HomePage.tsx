import { useEffect } from "react";
import { SearchInput } from "./SearchInput";
import { Results } from "./Results";
import styles from "./HomePage.module.css";
import { useTranslator } from "../hooks/useTranslator";

export function HomePage() {
  const {
    response,
    error,
    loading,
    initialQuery,
    handleTranslate,
  } = useTranslator();

  useEffect(() => {
    if (initialQuery) {
      handleTranslate(initialQuery);
    }
  }, []);

  return (
    <div className={styles.page}>
      <h1 className={styles.title}>wakaru</h1>
      <div className={styles.content}>
        <SearchInput onSubmit={handleTranslate} loading={loading} />
        <Results response={response} error={error} />
      </div>
    </div>
  );
}
