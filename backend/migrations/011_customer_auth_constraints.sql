SET search_path TO public;

ALTER TABLE otp_challenges
    ADD CONSTRAINT otp_challenges_verified_requires_consumed
    CHECK (verified_at IS NULL OR consumed_at IS NOT NULL);
