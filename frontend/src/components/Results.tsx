import type { TranslateResponse } from "../types";
import { TokenCard } from "./TokenCard.tsx"
import styles from './Results.module.css'

interface ResultsProps {
  response: TranslateResponse | null;
  error: string | null;
  examplesLoading?: boolean;
}

export function Results({ response, error, examplesLoading }: ResultsProps) {
  if (error) {
    return <div className={styles.error}>{error}</div>;
  }
  if (!response) {
    return null;
  }

  const tokensWithEntries = response.results
    .map((token, i) => ({ token, originalIndex: i }))
    .filter(({ token }) => token.entries && token.entries.length > 0);

  return (
    <div className={styles.results}>
      <div className={styles.displayString}>{response.displayString}</div>

      {tokensWithEntries.map(({ token, originalIndex }) => (
        <TokenCard token={token} index={originalIndex} key={originalIndex} examplesLoading={examplesLoading && !token.examples} />
      ))}
    </div>
  );
}
