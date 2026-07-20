import { useScroll } from "../hooks/useScroll";
import type { TranslateResponse } from "../types";
import { Entry } from "./Entry";
import styles from "./TokenCard.module.css"


export function TokenCard({ token }: { token: TranslateResponse["results"][number] }) {
  const { index, scrollForwards, scrollBackwards } = useScroll(token.entries);

  return (
    <div className={styles.tokenCard}>
      <Entry entry={token.entries[index]} />

      {token.examples && token.examples.length > 0 && (
        <ul className={styles.examples}>
          {token.examples.map((ex) => (
            <li key={ex.id}>{ex.text}</li>
          ))}
        </ul>
      )}

      <button onClick={scrollBackwards}>
        {"<"}
      </button>

      <button onClick={scrollForwards}>
        {">"}
      </button>

      <span>{index + 1}/{token.entries.length}</span>
    </div>
  );
}
