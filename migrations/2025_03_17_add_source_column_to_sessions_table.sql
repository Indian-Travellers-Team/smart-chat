-- Check if the 'source' column already exists in the 'sessions' table
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'sessions' 
          AND column_name = 'source'
    ) THEN
        ALTER TABLE "sessions"
        ADD COLUMN "source" VARCHAR(15) NOT NULL DEFAULT 'website';

        -- Update all existing records to set the value of "source" to 'website'
        UPDATE "sessions"
        SET "source" = 'website'
        WHERE "source" IS NULL;
    END IF;
END $$;
