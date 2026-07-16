export interface TranslateResponse {
  displayString: string;
  results: TokenResult[];
}

export interface TokenResult {
  entries: Entry[];
  examples: Example[];
}

export interface Entry {
  id: string;
  kanji: string[];
  kana: string[];
  translations: Translation[];
}

export interface Translation {
  senseId: number;
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
