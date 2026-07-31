import { Link } from "react-router-dom";
import { Entry } from "./Entry";

import styles from "./WordInfoPage.module.css";
import { useWordInfo } from "../hooks/useWordInfo";
import { useGenerateAIExamples } from "../hooks/useGenerateAIExamples";

export function WordInfoPage() {
  const {
    word,
    examples,
    searchQuery,
    entries,
    loading,
    error,
  } = useWordInfo();

  const {
    aiExamples,
    aiLoading,
    aiError,
    handleGenerateAIExamples,
  } = useGenerateAIExamples({ word })

  return (
    <div className={styles.page}>
      <Link to={
        searchQuery
          ? `/?q=${encodeURIComponent(searchQuery)}`
          : "/"
      } className={styles.backLink}>←Back to results</Link>

      {loading && <div className={styles.loading}>Loading...</div>}
      {error && <div className={styles.error}>Error: {error}</div>}

      {!loading && !error && entries.length === 0 && (
        <div className={styles.empty}>No entries found for "{word}"</div>
      )}

      {entries.map((entry) => (
        <div key={entry.id} className={styles.entrySection}>
          <Entry entry={entry} />
        </div>
      ))}

      {!loading && !error && entries.length > 0 && (
        <div className={styles.examplesSection}>
          <h3 className={styles.sectionTitle}>Examples</h3>
          <button
            className={styles.aiButton}
            onClick={handleGenerateAIExamples}
            disabled={aiLoading}
          >
            {aiLoading
              ? "Generating..."
              : aiExamples.length > 0
                ? "Regenerate AI Examples"
                : "Generate AI Examples"}
          </button>

          {aiError && <div className={styles.aiError}>{aiError}</div>}

          {aiExamples.length > 0 && (
            <ul className={styles.exampleList}>
              {aiExamples.map((ex, i) => (
                <li key={i} className={styles.exampleItem}>
                  <span className={styles.aiExampleStyle}>{ex.style}: </span>
                  {ex.text}
                </li>
              ))}
            </ul>
          )}

          {examples.length > 0 && (
            <ul className={styles.exampleList}>
              {examples.map((ex) => (
                <li key={ex.id} className={styles.exampleItem}>{ex.text}</li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
