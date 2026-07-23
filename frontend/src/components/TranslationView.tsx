import type { Translation } from "../types";
import { formatPos } from "../utils/posMap";
import styles from "./TranslationView.module.css"


interface TranslationViewProps {
  pos: string;
  senses: Translation[];
}

export const TranslationView = ({ pos, senses }: TranslationViewProps) => (
  <div key={pos} className={styles.senseGroup}>
    {pos && <div className={styles.posLabel}>{formatPos(pos)}</div>}
    <ol className={styles.senseList} start={1}>
      {senses.map((t) => (
        <li key={t.senseId} className={styles.sense}>
          <span className={styles.gloss}>
            {t.glosses.map(g => g.text).join(", ")}
          </span>
        </li>
      ))}
    </ol>
  </div>
)
