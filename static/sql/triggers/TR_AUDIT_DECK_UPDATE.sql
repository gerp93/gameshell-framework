CREATE
OR REPLACE TRIGGER TR_AUDIT_DECK_UPDATE
BEFORE UPDATE ON DECK
FOR EACH ROW
BEGIN
    IF NOT OLD.IS_HIDDEN THEN
        INSERT INTO AUDIT_DECK(
            AUDIT_TYPE,
            DECK_ID,
            NAME,
            PASSWORD_HASH,
            IS_PUBLIC_READONLY,
            IS_HIDDEN
        )
        VALUES (
            'UPDATE',
            OLD.ID,
            OLD.NAME,
            OLD.PASSWORD_HASH,
            OLD.IS_PUBLIC_READONLY,
            OLD.IS_HIDDEN
        );
    END
    IF;
END;
