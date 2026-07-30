import { useEffect } from "react";
import { SearchInput } from "./SearchInput";
import { Results } from "./Results";
import styles from "./HomePage.module.css";
import { useTranslator } from "../context/TranslatorContext";
import { useInitialQuery } from "../hooks/useInitialQuery";

export function HomePage() {
  const {
    response,
    error,
    loading,
    examplesLoading,
    handleTranslate,
  } = useTranslator();
  const { initialQuery } = useInitialQuery();

  useEffect(() => {
    if (initialQuery) {
      handleTranslate(initialQuery);
    }
  }, [initialQuery]);

  return (
    <div className={styles.page}>
      <h1 className={styles.title}>wakaru</h1>
      <div className={styles.content}>
        <SearchInput onSubmit={handleTranslate} loading={loading} />
        <Results response={response} error={error} examplesLoading={examplesLoading} />
      </div>
    </div>
  );
}
