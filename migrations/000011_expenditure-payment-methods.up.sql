INSERT INTO payment_method (display_name, owner)
    SELECT DISTINCT method, owner FROM expenditure
        WHERE method IS NOT NULL
            AND method != ''
            AND NOT EXISTS (
                SELECT 1 FROM payment_method
                WHERE display_name = method AND owner = expenditure.owner
            );

UPDATE expenditure
    SET method = (
        SELECT id FROM payment_method
        WHERE display_name = method AND owner = expenditure.owner
    )
    WHERE method IS NOT NULL;

ALTER TABLE IF EXISTS expenditure
    ALTER COLUMN method TYPE uuid USING method::uuid;
