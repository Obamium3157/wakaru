import type { Entry } from "../types"
import { buildRubySegments } from "../utils/furigana"
import styles from './Entry.module.css'

interface EntryProps {
  entry: Entry
}

export function Entry({ entry }: EntryProps) {
  const kanjiStr = entry.kanji?.[0] ?? "";
  const kanaStr = entry.kana?.[0] ?? "";
  const segments = entry.ruby ?? buildRubySegments(kanjiStr, kanaStr);

  console.log("Translations: ", entry.translations)

  return (
    <div className={styles.entry} key={entry.id}>
      <div className={styles.entryHeader}>
        {segments.map((seg, i) =>
          seg.reading ? (
            <ruby key={i} className={styles.kanji}>
              {seg.text}<rt className={styles.reading}>{seg.reading}</rt>
            </ruby>
          ) : (
            <span key={i}>
              {seg.text}
            </span>
          )
        )}
      </div>

      <ol className={styles.translations}>
        {entry.translations.map((t) => (
          <li className={styles.sense} key={t.senseId}>
            <span className={styles.gloss}>
              {t.glosses.map(g => g.text).join(", ")}
            </span>
          </li>
        ))}
      </ol>
    </div>
  )
}
