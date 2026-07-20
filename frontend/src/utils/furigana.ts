export interface RubySegment {
  text: string;
  reading?: string;
}

function codePoint(char: string): number {
  return char.codePointAt(0) ?? 0;
}

function isKanji(char: string): boolean {
  const cp = codePoint(char);
  return (cp >= 0x4E00 && cp <= 0x9FFF) || (cp >= 0x3400 && cp <= 0x4DBF);
}

function isHiragana(char: string): boolean {
  const cp = codePoint(char);
  return cp >= 0x3040 && cp <= 0x309F;
}

function isKatakana(char: string): boolean {
  const cp = codePoint(char);
  return cp >= 0x30A0 && cp <= 0x30FF;
}

function isKana(char: string): boolean {
  return isHiragana(char) || isKatakana(char);
}

function hasKanji(str: string): boolean {
  for (const char of str) {
    if (isKanji(char)) {
      return true;
    }
  }
  return false;
}

function detectOkurigana(str: string): { leading: string; trailing: string } {
  let leading = "";
  for (const char of str) {
    if (isKana(char)) {
      leading += char;
    } else {
      break;
    }
  }

  let trailing = "";
  const chars = [...str].reverse();
  for (const char of chars) {
    if (isKana(char)) {
      trailing = char + trailing;
    } else {
      break;
    }
  }

  return { leading, trailing };
}

interface ScriptSegment {
  text: string;
  isKanji: boolean;
}

function segmentByScript(str: string): ScriptSegment[] {
  const segments: ScriptSegment[] = [];
  let current = "";
  let currentIsKanji: boolean | null = null;

  for (const char of str) {
    const charIsKanji = isKanji(char);
    if (currentIsKanji !== null && currentIsKanji !== charIsKanji) {
      segments.push({ text: current, isKanji: currentIsKanji });
      current = "";
    }
    current += char;
    currentIsKanji = charIsKanji;
  }
  if (current) {
    segments.push({ text: current, isKanji: currentIsKanji! });
  }

  return segments;
}

function distributeReading(groups: string[], reading: string): string[] {
  if (groups.length === 0) {
    return [];
  }
  if (groups.length === 1) {
    return [reading];
  }

  const totalKanji = groups.reduce((sum, g) => sum + g.length, 0);
  const readings: string[] = [];
  let kanaIndex = 0;

  for (let i = 0; i < groups.length; i++) {
    const group = groups[i];
    const isLast = i === groups.length - 1;

    if (isLast) {
      readings.push(reading.slice(kanaIndex));
    } else {
      const proportion = group.length / totalKanji;
      const kanaCount = Math.round(proportion * reading.length);
      readings.push(reading.slice(kanaIndex, kanaIndex + kanaCount));
      kanaIndex += kanaCount;
    }
  }

  return readings;
}

export function buildRubySegments(kanjiStr: string, kanaStr: string): RubySegment[] {
  if (!kanjiStr || !kanaStr) {
    return [{ text: kanjiStr || kanaStr || "" }];
  }

  if (!hasKanji(kanjiStr)) {
    return [{ text: kanjiStr }];
  }

  const { leading, trailing } = detectOkurigana(kanjiStr);
  const kanjiReading = kanaStr
    .slice(leading.length, kanaStr.length - (trailing.length || 0));

  const pureKanji = kanjiStr.slice(
    leading.length,
    kanjiStr.length - (trailing.length || 0)
  );

  const pureSegments = segmentByScript(pureKanji);
  const kanjiOnly = pureSegments
    .filter((s) => s.isKanji)
    .map((s) => s.text);
  const readings = distributeReading(kanjiOnly, kanjiReading);

  const segments: RubySegment[] = [];

  if (leading) {
    segments.push({ text: leading });
  }

  let kanjiIdx = 0;
  for (const seg of pureSegments) {
    if (seg.isKanji) {
      segments.push({ text: seg.text, reading: readings[kanjiIdx] || "" });
      kanjiIdx++;
    } else {
      segments.push({ text: seg.text });
    }
  }

  if (trailing) {
    segments.push({ text: trailing });
  }

  return segments;
}
