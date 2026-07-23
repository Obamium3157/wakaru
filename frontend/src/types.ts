export interface TranslateResponse {
  displayString: string;
  results: TokenResult[];
}

export interface TokenResult {
  entries: Entry[];
  examples: Example[];
  posMajor: string;
}

export interface Entry {
  id: string;
  kanji: string[];
  kana: string[];
  ruby?: RubySegment[];
  translations: Translation[];
}

export interface RubySegment {
  text: string;
  reading?: string;
}

export interface Translation {
  senseId: number;
  pos?: string;
  glosses: Gloss[];
}

export interface Gloss {
  lang: string;
  text: string;
}

export interface Example {
  id: number;
  text: string;
}

export interface AddBasicNoteRequest {
  deckName: string;
  front: string;
  back: string;
  tags?: string[];
}
