ALTER TABLE IF EXISTS expenditure
    ALTER COLUMN method TYPE VARCHAR(50) USING method::text;

UPDATE expenditure
SET method = (
    SELECT display_name FROM payment_method
    WHERE id = method::uuid AND owner = expenditure.owner
)
WHERE method IS NOT NULL AND METHOD != '';
