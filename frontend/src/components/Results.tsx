import type { TranslateResponse } from "../types";
import { TokenCard } from "./TokenCard.tsx"
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
        <TokenCard token={token} key={i} />
      ))}
    </div>
  );
}
