import { useState } from "react";
import { SearchInput } from "./components/SearchInput";
import { Results } from "./components/Results";
import { translate } from "./api";
import type { TranslateResponse } from "./types";
import styles from './App.module.css'

function App() {
  const [response, setResponse] = useState<TranslateResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleTranslate(text: string) {
    setLoading(true);
    setError(null);

    try {
      const data = await translate(text);
      setResponse(data);
    } catch (e) {
      setResponse(null);
      setError(e instanceof Error ? e.message : "unknown error");
    } finally {
      setLoading(false);
    }
  }

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
