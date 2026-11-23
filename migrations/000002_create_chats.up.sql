BEGIN;

CREATE TABLE chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind TEXT NOT NULL,
    title TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    direct_user_low UUID REFERENCES users(id),
    direct_user_high UUID REFERENCES users(id),

    CONSTRAINT chats_kind_valid CHECK (
        kind IN ('direct', 'group')
    ),

    CONSTRAINT chats_shape_valid CHECK (
        (
            kind = 'direct'
            AND title IS NULL
            AND direct_user_low IS NOT NULL
            AND direct_user_high IS NOT NULL
            AND direct_user_low < direct_user_high
            AND created_by IN (direct_user_low, direct_user_high)
        )
        OR
        (
            kind = 'group'
            AND title IS NOT NULL
            AND title = btrim(title)
            AND char_length(title) BETWEEN 1 AND 100
            AND direct_user_low IS NULL
            AND direct_user_high IS NULL
        )
    )
);

CREATE UNIQUE INDEX chats_direct_pair_unique
    ON chats (direct_user_low, direct_user_high)
    WHERE kind = 'direct';

CREATE TABLE chat_members (
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    role TEXT NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (chat_id, user_id),

    CONSTRAINT chat_members_role_valid CHECK (
        role IN ('admin', 'member')
    )
);

CREATE INDEX chat_members_user_chats
    ON chat_members (user_id, chat_id);

CREATE UNIQUE INDEX chat_members_single_admin
    ON chat_members (chat_id)
    WHERE role = 'admin';

COMMIT;