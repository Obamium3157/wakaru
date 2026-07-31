package repository

const findQuery = `
SELECT word_id FROM kanji WHERE text = ?
UNION
SELECT word_id FROM kana WHERE text = ?
ORDER BY word_id;
`

const findFilteredQuery = `
SELECT DISTINCT matches.word_id
FROM (
	SELECT word_id FROM kanji WHERE text = ?
  UNION
  SELECT word_id FROM kana WHERE text = ?
) AS matches
WHERE EXISTS (
  SELECT 1
  FROM sense s
  JOIN sense_part_of_speech sps
	  ON sps.sense_id = s.id
  JOIN tag t ON t.id = sps.tag_id
    WHERE s.word_id = matches.word_id
      AND t.label IN (%s)
)
ORDER BY matches.word_id;
`

const findByKanjiQuery = `
SELECT word_id
FROM kanji
WHERE text = ?
ORDER BY word_id;
`

const findByKanaQuery = `
SELECT word_id
FROM kana
WHERE text = ?
ORDER BY word_id;
`

const findKanjiBatchQuery = `
SELECT word_id, text
FROM kanji
WHERE word_id IN (%s)
ORDER BY word_id, display_order;
`

const findKanaBatchQuery = `
SELECT word_id, text
FROM kana
WHERE word_id IN (%s)
ORDER BY word_id, display_order;
`

const findTranslationsBatchQuery = `
SELECT
  s.word_id,
  s.id,
  g.lang,
  g.text,
  pos.tags AS pos
FROM sense s
JOIN gloss g ON g.sense_id = s.id
LEFT JOIN (
  SELECT
    sps.sense_id,
    GROUP_CONCAT(t.label, ', ') AS tags
  FROM sense_part_of_speech sps
  JOIN tag t
	  ON t.id = sps.tag_id
  WHERE sps.sense_id IN (SELECT id FROM sense WHERE word_id IN (%s))
  GROUP BY sps.sense_id
) pos ON pos.sense_id = s.id
WHERE s.word_id IN (%s)
ORDER BY s.word_id, s.display_order, g.display_order;
`

const findAllFormsQuery = `
SELECT text FROM kanji
UNION
SELECT text FROM kana;
`

const findKanaReadingsForKanjiBatchQuery = `
SELECT kn.word_id, kap.kanji_text, kn.text AS kana_text
FROM kana kn
JOIN kana_applies_to_kanji kap
  ON kap.kana_id = kn.id
WHERE kn.word_id IN (%s)
ORDER BY kn.word_id, kn.display_order, kap.display_order;
`
