import { Link, useSearchParams } from "react-router-dom";
import { useScroll } from "../hooks/useScroll";
import type { TranslateResponse } from "../types";
import { Entry } from "./Entry";
import styles from "./TokenCard.module.css"

export function TokenCard({ token, index }: { token: TranslateResponse["results"][number]; index: number }) {
  const { index: entryIndex, scrollForwards, scrollBackwards } = useScroll(token.entries);
  const entry = token.entries[entryIndex];
  const wordParam = encodeURIComponent(entry.kanji?.[0] ?? entry.kana?.[0] ?? "");
  const [searchParams] = useSearchParams();
  const searchQuery = searchParams.get("q") ?? "";

  return (
    <div className={styles.tokenCard} id={`token-${index}`}>
      <Entry entry={entry} />

      <div className={styles.footer}>
        <Link
          to={`/info/${wordParam}?pos=${token.posMajor}`}
          state={{ examples: token.examples, searchQuery }}
          className={styles.moreDetails}
        >
          More details...
        </Link>

        {token.entries.length > 1 && (
          <div className={styles.entryNav}>
            <button onClick={scrollBackwards} className={styles.navBtn}>{"<"}</button>
            <span className={styles.entryCount}>{entryIndex + 1}/{token.entries.length}</span>
            <button onClick={scrollForwards} className={styles.navBtn}>{">"}</button>
          </div>
        )}
      </div>
    </div>
  );
}
