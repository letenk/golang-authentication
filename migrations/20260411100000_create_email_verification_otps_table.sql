-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS email_verification_otps (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       VARCHAR(10) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at    TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_email_verification_otps_user_id ON email_verification_otps(user_id);
CREATE INDEX IF NOT EXISTS idx_email_verification_otps_expires_at ON email_verification_otps(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS email_verification_otps CASCADE;
-- +goose StatementEnd
