-- 000005_mfa_recovery_codes.up.sql — one-time MFA recovery codes.
-- Only password-hasher outputs are stored; raw recovery codes are shown once
-- after enrollment and are consumed on successful use.
ALTER TABLE credentials
    ADD COLUMN recovery_code_hashes text[] NOT NULL DEFAULT '{}';
