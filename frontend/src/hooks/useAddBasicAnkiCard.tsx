import { addBasicNote } from "../api";
import type { AddBasicNoteRequest } from "../types";

interface CreateBasicAnkiCardProps {
  front: string;
  back: string;
}

export function useAddBasicAnkiCard() {
  const addBasicAnkiCard = ({ front, back }: CreateBasicAnkiCardProps) => {
    const request: AddBasicNoteRequest = {
      deckName: "test deck",
      front: front,
      back: back,
      tags: ["test"]
    }

    addBasicNote(request)
  }

  return {
    addBasicAnkiCard,
  }
}
