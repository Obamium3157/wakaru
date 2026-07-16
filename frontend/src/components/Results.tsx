import type { TranslateResponse } from "../types";
import styles from './Results.module.css'

interface ResultsProps {
  response: TranslateResponse | null;
  error: string | null;
}

export function Results({ response, error }: ResultsProps) {
  if (error) {
    return <div className={styles.error}>{error}</div>;
  }
  if (!response) {
    return null;
  }

  const tokensWithEntries = response.results.filter(
    (r) => r.entries && r.entries.length > 0
  );

  return (
    <div className={styles.results}>
      <div className={styles.displayString}>{response.displayString}</div>

      {tokensWithEntries.map((token, i) => (
        <div className={styles.tokenCard} key={i}>
          {token.entries.map((entry) => (
            <div className={styles.entry} key={entry.id}>
              <div className={styles.entryHeader}>
                <span className={styles.kanji}>{(entry.kanji ?? []).join(", ")}</span>
                <span className={styles.kana}>{(entry.kana ?? []).join(", ")}</span>
              </div>

              <div className={styles.translations}>
                {entry.translations.map((t) => (
                  <div className={styles.sense} key={t.senseId}>
                    <div className={styles.glosses}>
                      {t.glosses.map((g, j) => (
                        <span className={styles.gloss} key={j}>
                          {g.text}
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}

          {token.examples && token.examples.length > 0 && (
            <ul className={styles.examples}>
              {token.examples.map((ex) => (
                <li key={ex.id}>{ex.text}</li>
              ))}
            </ul>
          )}
        </div>
      ))}
    </div>
  );
}
