package repository

import "unicode"

func isKanji(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		(r >= 0x3400 && r <= 0x4DBF)
}

func toHiragana(s string) string {
	var out []rune
	for _, r := range s {
		if r >= 0x30A1 && r <= 0x30F4 {
			out = append(out, r-0x60)
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

func hasKanji(s string) bool {
	for _, r := range s {
		if isKanji(r) {
			return true
		}
	}
	return false
}

type scriptSegment struct {
	text    string
	isKanji bool
}

func segmentByScript(s string) []scriptSegment {
	var segments []scriptSegment
	var current []rune
	var currentIsKanji *bool

	for _, r := range s {
		rIsKanji := isKanji(r)
		if currentIsKanji != nil && *currentIsKanji != rIsKanji {
			segments = append(segments, scriptSegment{
				text:    string(current),
				isKanji: *currentIsKanji,
			})
			current = nil
		}
		current = append(current, r)
		currentIsKanji = &rIsKanji
	}
	if len(current) > 0 {
		segments = append(segments, scriptSegment{
			text:    string(current),
			isKanji: *currentIsKanji,
		})
	}

	return segments
}

func BuildRubySegments(surface, reading string) []RubySegment {
	if surface == "" || reading == "" {
		return []RubySegment{{Text: surface}}
	}

	if !hasKanji(surface) {
		return []RubySegment{{Text: surface}}
	}

	segments := segmentByScript(surface)

	if !hasKanaSegment(segments) {
		return []RubySegment{{Text: surface, Reading: &reading}}
	}

	if result, ok := alignReading(segments, reading); ok {
		return result
	}

	return []RubySegment{{Text: surface, Reading: &reading}}
}

func hasKanaSegment(segments []scriptSegment) bool {
	for _, seg := range segments {
		if !seg.isKanji {
			return true
		}
	}
	return false
}

func alignReading(segments []scriptSegment, reading string) ([]RubySegment, bool) {
	r := []rune(reading)
	rNorm := []rune(toHiragana(reading))
	pos := 0
	result := make([]RubySegment, 0, len(segments))

	for i, seg := range segments {
		if seg.isKanji {
			part, next, ok := kanjiReading(segments, i, r, rNorm, pos)
			if !ok {
				return nil, false
			}
			result = append(result, RubySegment{Text: seg.text, Reading: &part})
			pos = next
		} else {
			next, ok := kanaReading(seg.text, r, rNorm, pos)
			if !ok {
				return nil, false
			}
			result = append(result, RubySegment{Text: seg.text})
			pos = next
		}
	}

	return result, pos == len(r)
}

func kanjiReading(segments []scriptSegment, i int, r, rNorm []rune, pos int) (part string, next int, ok bool) {
	if i == len(segments)-1 {
		part = string(r[pos:])
		if len(part) == 0 {
			return "", pos, false
		}
		return part, len(r), true
	}

	nextKanaNorm := toHiragana(segments[i+1].text)
	idx := indexOfRune(rNorm[pos:], []rune(nextKanaNorm))
	if idx <= 0 {
		return "", pos, false
	}
	return string(r[pos : pos+idx]), pos + idx, true
}

func kanaReading(text string, r, rNorm []rune, pos int) (int, bool) {
	k := []rune(text)
	kNorm := []rune(toHiragana(text))
	if pos+len(k) > len(r) || string(rNorm[pos:pos+len(k)]) != string(kNorm) {
		return pos, false
	}
	return pos + len(k), true
}

func indexOfRune(s, substr []rune) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := range substr {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
