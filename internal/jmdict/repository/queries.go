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

const findKanjiQuery = `
SELECT text
FROM kanji
WHERE word_id = ?
ORDER BY display_order;
`

const findKanaQuery = `
SELECT text
FROM kana
WHERE word_id = ?
ORDER BY display_order;
`

const findTranslationsQuery = `
SELECT
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
  GROUP BY sps.sense_id
) pos ON pos.sense_id = s.id
WHERE s.word_id = ?
ORDER BY s.display_order, g.display_order;
`

const findAllFormsQuery = `
SELECT text FROM kanji
UNION
SELECT text FROM kana;
`

const findKanaReadingsForKanjiQuery = `
SELECT kap.kanji_text, kn.text AS kana_text
FROM kana kn
JOIN kana_applies_to_kanji kap
  ON kap.kana_id = kn.id
WHERE kn.word_id = ?
ORDER BY kn.display_order, kap.display_order;
`
