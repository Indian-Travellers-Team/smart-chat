DO $$
BEGIN
    IF to_regclass('public.auth_user_conversation_links') IS NOT NULL
       AND to_regclass('public.auth_user_conversation') IS NULL THEN
        ALTER TABLE public.auth_user_conversation_links RENAME TO auth_user_conversation;
    END IF;
END $$;