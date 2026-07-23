import { useGroupedByPos } from "../hooks/useGroupedByPos";
import type { Entry as EntryType } from "../types"
import { buildRubySegments } from "../utils/furigana"
import styles from './Entry.module.css'
import { TranslationView } from "./TranslationView";
import { WordDisplay } from "./WordDisplay";

interface EntryProps {
  entry: EntryType;
}

export function Entry({ entry }: EntryProps) {
  const kanjiStr = entry.kanji?.[0] ?? "";
  const kanaStr = entry.kana?.[0] ?? "";
  const segments = entry.ruby ?? buildRubySegments(kanjiStr, kanaStr);
  const groupedByPos = useGroupedByPos({ translations: entry.translations });

  return (
    <div className={styles.entry}>
      <WordDisplay segments={segments} />

      <div className={styles.translations}>
        {Object.entries(groupedByPos).map(([pos, senses]) => (
          <TranslationView pos={pos} senses={senses} />
        ))}
      </div>
    </div>
  )
}
