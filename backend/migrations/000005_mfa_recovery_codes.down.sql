-- 000005_mfa_recovery_codes.down.sql
ALTER TABLE credentials
    DROP COLUMN IF EXISTS recovery_code_hashes;
