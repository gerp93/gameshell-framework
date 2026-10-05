CREATE
OR REPLACE TRIGGER TR_AUDIT_DECK_DELETE
BEFORE DELETE ON DECK
FOR EACH ROW
BEGIN
    -- Games own their own CARD tables; card auditing on deck delete happens in
    -- the game's OnDeckDeleting hook (FK cascade does not fire card triggers).
    INSERT INTO AUDIT_DECK(
        AUDIT_TYPE,
        DECK_ID,
        NAME,
        PASSWORD_HASH,
        IS_PUBLIC_READONLY
    )
    VALUES (
        'DELETE',
        OLD.ID,
        OLD.NAME,
        OLD.PASSWORD_HASH,
        OLD.IS_PUBLIC_READONLY
    );
END;
