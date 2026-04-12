DO $$
BEGIN
    IF to_regclass('public.auth_user_conversation') IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = 'auth_user_conversation'
              AND column_name = 'started'
        ) THEN
            ALTER TABLE public.auth_user_conversation
                ADD COLUMN started BOOLEAN NOT NULL DEFAULT FALSE;
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = 'auth_user_conversation'
              AND column_name = 'resolved'
        ) THEN
            ALTER TABLE public.auth_user_conversation
                ADD COLUMN resolved BOOLEAN NOT NULL DEFAULT FALSE;
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = 'auth_user_conversation'
              AND column_name = 'comments'
        ) THEN
            ALTER TABLE public.auth_user_conversation
                ADD COLUMN comments TEXT NOT NULL DEFAULT '';
        END IF;
    END IF;
END $$;