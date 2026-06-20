ALTER TABLE user_profile
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS email_reminders_enabled;
