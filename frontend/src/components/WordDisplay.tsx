import type { RubySegment } from "../types";
import styles from "./WordDisplay.module.css"

interface WordDisplayProps {
  segments: RubySegment[];
}

export const WordDisplay = ({ segments }: WordDisplayProps) => (
  <div className={styles.header}>
    {segments.map((seg, i) =>
      seg.reading ? (
        <ruby key={i} className={styles.kanji}>
          {seg.text}<rt className={styles.reading}>{seg.reading}</rt>
        </ruby>
      ) : (
        <span key={i}>{seg.text}</span>
      )
    )}
  </div>
)
