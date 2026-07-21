import { useEffect } from "react";
import { SearchInput } from "./components/SearchInput";
import { Results } from "./components/Results";
import styles from './App.module.css'
import { useTranslator } from "./hooks/useTranslator";

function App() {
  const {
    response,
    error,
    loading,
    initialQuery,
    handleTranslate
  } = useTranslator()

  useEffect(() => {
    if (initialQuery) {
      handleTranslate(initialQuery);
    }
  }, []);


  return (
    <div className={styles.app}>
      <h1>wakaru</h1>
      <div className={styles.content}>
        <SearchInput onSubmit={handleTranslate} loading={loading} />
        <Results response={response} error={error} />
      </div>
    </div>
  );
}

export default App;
