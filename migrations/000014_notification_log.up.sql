CREATE TABLE notification_log (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID        NOT NULL REFERENCES user_profile(id) ON DELETE CASCADE,
    payment_method_id UUID        NOT NULL REFERENCES payment_method(id) ON DELETE CASCADE,
    renewal_year      INT         NOT NULL,
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (payment_method_id, renewal_year)
);
