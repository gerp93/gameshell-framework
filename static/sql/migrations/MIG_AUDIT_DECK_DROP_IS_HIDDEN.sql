-- Drops the retired AUDIT_DECK.IS_HIDDEN column. Idempotent (see
-- MIG_DECK_DROP_IS_HIDDEN.sql). Must run after the deck audit triggers are
-- replaced so no trigger inserts into the dropped column.
ALTER TABLE AUDIT_DECK DROP COLUMN IF EXISTS IS_HIDDEN;
