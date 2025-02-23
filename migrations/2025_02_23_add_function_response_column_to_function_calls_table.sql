-- Check if the 'function_response' column already exists in the 'function_calls' table
DO $$
BEGIN
    -- If the column doesn't exist, add it
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'function_calls' AND column_name = 'function_response') THEN
        -- Add the 'function_response' column with the type 'varchar'
        ALTER TABLE "function_calls"
        ADD COLUMN "function_response" VARCHAR(500);

        -- Optionally, update all existing records to set the value of "function_response" to a default value (if necessary)
        UPDATE "function_calls"
        SET "function_response" = 'default response'  -- Replace 'default response' with your desired default value
        WHERE "function_response" IS NULL;
    END IF;
END $$;
