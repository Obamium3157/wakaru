PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS `tag` (
  `id` INTEGER PRIMARY KEY,
  `label` TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS `word` (
  `id` TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS `kana` (
  `id` INTEGER PRIMARY KEY,
  `word_id` TEXT NOT NULL REFERENCES `word` (`id`) ON DELETE CASCADE,
  `is_common` INTEGER NOT NULL DEFAULT 0 CHECK (`is_common` IN (0, 1)),
  `text` TEXT NOT NULL,
  `display_order` INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS `idx_kana_word_id` ON `kana` (`word_id`);
CREATE INDEX IF NOT EXISTS `idx_kana_text` ON `kana` (`text`);
CREATE INDEX IF NOT EXISTS `idx_kana_word_order` ON `kana` (`word_id`, `display_order`);

CREATE TABLE IF NOT EXISTS `kana_tag` (
  `kana_id` INTEGER NOT NULL REFERENCES `kana` (`id`) ON DELETE CASCADE,
  `tag_id` INTEGER NOT NULL REFERENCES `tag` (`id`),

  PRIMARY KEY (`kana_id`, `tag_id`)
);

CREATE TABLE IF NOT EXISTS `kana_applies_to_kanji` (
  `kana_id` INTEGER NOT NULL REFERENCES `kana` (`id`) ON DELETE CASCADE,
  `kanji_text` TEXT NOT NULL,

  PRIMARY KEY (`kana_id`, `kanji_text`)
);

CREATE TABLE IF NOT EXISTS `kanji` (
  `id` INTEGER PRIMARY KEY,
  `word_id` TEXT NOT NULL REFERENCES `word` (`id`) ON DELETE CASCADE,
  `is_common` INTEGER NOT NULL DEFAULT 0 CHECK (`is_common` IN (0, 1)),
  `text` TEXT NOT NULL,
  `display_order` INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS `idx_kanji_word_id` ON `kanji` (`word_id`);
CREATE INDEX IF NOT EXISTS `idx_kanji_text` ON `kanji` (`text`);
CREATE INDEX IF NOT EXISTS `idx_kanji_word_order` ON `kanji` (`word_id`, `display_order`);

CREATE TABLE IF NOT EXISTS `kanji_tag` (
  `kanji_id` INTEGER NOT NULL REFERENCES `kanji` (`id`) ON DELETE CASCADE,
  `tag_id` INTEGER NOT NULL REFERENCES `tag` (`id`),

  PRIMARY KEY (`kanji_id`, `tag_id`)
);

CREATE TABLE IF NOT EXISTS `sense` (
  `id` INTEGER PRIMARY KEY,
  `word_id` TEXT NOT NULL REFERENCES `word` (`id`) ON DELETE CASCADE,
  `display_order` INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS `idx_sense_word_id` ON `sense` (`word_id`);
CREATE INDEX IF NOT EXISTS `idx_sense_word_order` ON `sense` (`word_id`, `display_order`);

CREATE TABLE IF NOT EXISTS `sense_dialect` (
  `sense_id` INTEGER NOT NULL REFERENCES `sense` (`id`) ON DELETE CASCADE,
  `tag_id` INTEGER NOT NULL REFERENCES `tag` (`id`),

  PRIMARY KEY (`sense_id`, `tag_id`)
);

CREATE TABLE IF NOT EXISTS `sense_field` (
  `sense_id` INTEGER NOT NULL REFERENCES `sense` (`id`) ON DELETE CASCADE,
  `tag_id` INTEGER NOT NULL REFERENCES `tag` (`id`),

  PRIMARY KEY (`sense_id`, `tag_id`)
);

CREATE TABLE IF NOT EXISTS `sense_misc` (
  `sense_id` INTEGER NOT NULL REFERENCES `sense` (`id`) ON DELETE CASCADE,
  `tag_id` INTEGER NOT NULL REFERENCES `tag` (`id`),

  PRIMARY KEY (`sense_id`, `tag_id`)
);

CREATE TABLE IF NOT EXISTS `sense_part_of_speech` (
  `sense_id` INTEGER NOT NULL REFERENCES `sense` (`id`) ON DELETE CASCADE,
  `tag_id` INTEGER NOT NULL REFERENCES `tag` (`id`),

  PRIMARY KEY (`sense_id`, `tag_id`)
);

CREATE TABLE IF NOT EXISTS `sense_applies_to_kana` (
  `sense_id` INTEGER NOT NULL REFERENCES `sense` (`id`) ON DELETE CASCADE,
  `kana_text` TEXT NOT NULL,

  PRIMARY KEY (`sense_id`, `kana_text`)
);

CREATE TABLE IF NOT EXISTS `sense_applies_to_kanji` (
  `sense_id` INTEGER NOT NULL REFERENCES `sense` (`id`) ON DELETE CASCADE,
  `kanji_text` TEXT NOT NULL,

  PRIMARY KEY (`sense_id`, `kanji_text`)
);

CREATE TABLE IF NOT EXISTS `sense_info` (
  `id` INTEGER PRIMARY KEY,
  `sense_id` INTEGER NOT NULL REFERENCES `sense` (`id`) ON DELETE CASCADE,
  `text` VARCHAR(255) NOT NULL,
  `display_order` INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS `idx_sense_info_sense_id` ON `sense_info` (`sense_id`);

CREATE TABLE IF NOT EXISTS `gloss` (
  `id` INTEGER PRIMARY KEY,
  `sense_id` INTEGER NOT NULL REFERENCES `sense` (`id`) ON DELETE CASCADE,
  `lang` VARCHAR(3) NOT NULL,
  `text` TEXT NOT NULL,
  `gloss_type` VARCHAR(20) CHECK (
    `gloss_type`
    IN ('literal', 'figurative', 'explanation', 'trademark')
    OR `gloss_type` IS null
  ),
  `gender` VARCHAR(20) CHECK (
    `gender`
    IN ('masculine', 'feminine', 'neuter')
    OR `gender` IS null
  ),
  `display_order` INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS `idx_gloss_sense_id` ON `gloss` (`sense_id`);
CREATE INDEX IF NOT EXISTS `idx_gloss_text` ON `gloss` (`text`);
CREATE INDEX IF NOT EXISTS `idx_gloss_display_order` ON `gloss` (`display_order`);

CREATE TABLE IF NOT EXISTS `language_source` (
  `id` INTEGER PRIMARY KEY,
  `sense_id` INTEGER NOT NULL REFERENCES `sense` (`id`) ON DELETE CASCADE,
  `lang` VARCHAR(3) NOT NULL,
  `text` TEXT,
  `is_full` INTEGER NOT NULL DEFAULT 1 CHECK (`is_full` IN (0, 1)),
  `is_wasei` INTEGER NOT NULL DEFAULT 0 CHECK (`is_wasei` IN (0, 1))
);

CREATE INDEX IF NOT EXISTS `idx_language_source_sense_id` ON `language_source` (`sense_id`);

CREATE TABLE IF NOT EXISTS `xref` (
  `id` INTEGER PRIMARY KEY,
  `sense_id` INTEGER NOT NULL REFERENCES `sense` (`id`) ON DELETE CASCADE,
  `relation_type` VARCHAR(10) NOT NULL CHECK (`relation_type` IN ('related', 'antonym')),
  `headword` TEXT NOT NULL,
  `reading` TEXT,
  `sense_index` INTEGER,
  `display_order` INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS `idx_xref_sense_id` ON `xref` (`sense_id`);
CREATE INDEX IF NOT EXISTS `idx_xref_headword` ON `xref` (`headword`);
CREATE INDEX IF NOT EXISTS `idx_xref_reading` ON `xref` (`reading`);
CREATE INDEX IF NOT EXISTS `idx_xref_display_order` ON `xref` (`display_order`);

CREATE TABLE IF NOT EXISTS `xref_resolved` (
  `xref_id` INTEGER PRIMARY KEY REFERENCES `xref` (`id`) ON DELETE CASCADE,
  `word_id` TEXT NOT NULL REFERENCES `word` (`id`) ON DELETE CASCADE,
  `sense_id` INTEGER REFERENCES `sense` (`id`) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS `idx_xref_resolved_word_id` ON `xref_resolved` (`word_id`);
CREATE INDEX IF NOT EXISTS `idx_xref_resolved_sense_id` ON `xref_resolved` (`sense_id`);
