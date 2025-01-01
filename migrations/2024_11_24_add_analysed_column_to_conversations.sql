-- Check if the 'analysed' column already exists in the 'conversations' table
DO $$
BEGIN
    -- If the column doesn't exist, add it
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'conversations' AND column_name = 'analysed') THEN
        -- Add the 'analysed' column with the default value 'TRUE'
        ALTER TABLE "conversations"
        ADD COLUMN "analysed" BOOLEAN NOT NULL DEFAULT TRUE;

        -- Update all existing records to set the value of "analysed" to TRUE
        UPDATE "conversations"
        SET "analysed" = TRUE
        WHERE "analysed" IS NULL;
    END IF;
END $$;
