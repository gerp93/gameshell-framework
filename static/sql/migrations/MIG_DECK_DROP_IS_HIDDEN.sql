-- Drops the retired DECK.IS_HIDDEN column from databases provisioned before
-- games stopped modeling internal decks as hidden DECK rows. Idempotent:
-- IF EXISTS makes it a no-op on fresh databases (the column is no longer in
-- DECK.sql). Must run AFTER the deck audit triggers are replaced, since the
-- old triggers referenced OLD.IS_HIDDEN.
ALTER TABLE DECK DROP COLUMN IF EXISTS IS_HIDDEN;
