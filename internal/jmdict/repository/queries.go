package repository

const findQuery = `
SELECT DISTINCT word_id
FROM (
    SELECT word_id
    FROM kanji
    WHERE text = ?

    UNION

    SELECT word_id
    FROM kana
    WHERE text = ?
)
ORDER BY word_id;
`

const findByKanjiQuery = `
SELECT DISTINCT word_id
FROM kanji
WHERE text = ?
ORDER BY word_id;
`

const findByKanaQuery = `
SELECT DISTINCT word_id
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
    g.text
FROM sense s
JOIN gloss g
ON g.sense_id = s.id
WHERE s.word_id = ?
ORDER BY
    s.display_order,
    g.display_order;
`

const findAllFormsQuery = `
SELECT DISTINCT text FROM kanji
UNION
SELECT DISTINCT text FROM kana;
`

const findKanaReadingsForKanjiQuery = `
SELECT kap.kanji_text, kn.text AS kana_text
FROM kana kn
JOIN kana_applies_to_kanji kap
    ON kap.kana_id = kn.id
WHERE kn.word_id = ?
ORDER BY kn.display_order, kap.display_order;
`
