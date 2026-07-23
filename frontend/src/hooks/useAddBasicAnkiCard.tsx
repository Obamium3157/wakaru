import { addBasicNote } from "../api";
import type { AddBasicNoteRequest, Entry, RubySegment } from "../types";

interface CreateBasicAnkiCardProps {
  entry: Entry;
  segments: RubySegment[];
}

export function useAddBasicAnkiCard({ entry, segments }: CreateBasicAnkiCardProps) {
  const addBasicAnkiCard = () => {
    const front = segments
      .map(s => s.reading ? `<ruby>${s.text}<rt>${s.reading}</rt></ruby>` : s.text)
      .join("");

    const glosses = entry.translations[0]?.glosses ?? [];
    if (glosses.length === 0) {
      window.alert("No translations available for this word.");
      return;
    }
    const back = glosses.map(g => g.text).join(", ");

    const request: AddBasicNoteRequest = {
      deckName: "test deck",
      front,
      back,
    };

    addBasicNote(request).then(() => {
      window.alert("Card created successfully.");
    });
  }

  return {
    addBasicAnkiCard,
  }
}
