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

	hasKanaAnchor := false
	for _, seg := range segments {
		if !seg.isKanji {
			hasKanaAnchor = true
			break
		}
	}

	if !hasKanaAnchor {
		return []RubySegment{{Text: surface, Reading: &reading}}
	}

	r := []rune(reading)
	rNorm := []rune(toHiragana(reading))
	pos := 0
	result := make([]RubySegment, 0, len(segments))
	valid := true

	for i, seg := range segments {
		if seg.isKanji {
			isLast := i == len(segments)-1

			if isLast {
				readingPart := string(r[pos:])
				if len(readingPart) == 0 {
					valid = false
					break
				}
				result = append(result, RubySegment{Text: seg.text, Reading: &readingPart})
				pos = len(r)
			} else {
				nextKanaNorm := toHiragana(segments[i+1].text)
				idx := indexOfRune(rNorm[pos:], []rune(nextKanaNorm))
				if idx < 0 {
					valid = false
					break
				}
				readingPart := string(r[pos : pos+idx])
				if len(readingPart) == 0 {
					valid = false
					break
				}
				result = append(result, RubySegment{Text: seg.text, Reading: &readingPart})
				pos += idx
			}
		} else {
			k := []rune(seg.text)
			kNorm := []rune(toHiragana(seg.text))
			if pos+len(k) > len(r) || string(rNorm[pos:pos+len(k)]) != string(kNorm) {
				valid = false
				break
			}
			result = append(result, RubySegment{Text: seg.text})
			pos += len(k)
		}
	}

	if !valid || pos != len(r) {
		return []RubySegment{{Text: surface, Reading: &reading}}
	}

	return result
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
