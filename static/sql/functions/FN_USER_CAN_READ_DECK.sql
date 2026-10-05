CREATE
OR REPLACE FUNCTION FN_USER_CAN_READ_DECK(
    IN VAR_USER_ID UUID,
    IN VAR_DECK_ID UUID
)
RETURNS BOOLEAN
BEGIN
    -- USER IS ADMIN
    IF EXISTS(SELECT ID FROM USER WHERE ID = VAR_USER_ID AND IS_ADMIN = 1) THEN
        RETURN 1;
    END
    IF;

    -- USER HAS ACCESS
    IF EXISTS(
        SELECT
            ID
        FROM USER_ACCESS_DECK
        WHERE USER_ID = VAR_USER_ID
            AND DECK_ID = VAR_DECK_ID
    ) THEN
        RETURN 1;
    END
    IF;

    -- DECK IS PUBLIC READ-ONLY: same condition SP_GET_READABLE_DECKS already
    -- uses to list it for every user. Deliberately a separate function from
    -- FN_USER_HAS_DECK_ACCESS rather than an added branch there: that
    -- function also gates deck/card WRITE actions (rename, delete, card
    -- CRUD), and public-readonly must never grant those.
    IF EXISTS(
        SELECT
            ID
        FROM DECK
        WHERE ID = VAR_DECK_ID
            AND IS_PUBLIC_READONLY = 1
    ) THEN
        RETURN 1;
    END
    IF;

    RETURN VAR_DECK_ID IS NULL;
END;
