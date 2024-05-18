-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    versions bigint not null default 0,

    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp null,

    created_by bigint not null,
    CONSTRAINT users_created_by_foreign FOREIGN KEY (created_by) REFERENCES users (id),

    updated_by bigint not null,
    CONSTRAINT users_updated_by_foreign FOREIGN KEY (updated_by) REFERENCES users (id),

    deleted_by bigint null default null,
    CONSTRAINT users_deleted_by_foreign FOREIGN KEY (deleted_by) REFERENCES users (id),

    name varchar(191)  NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT '0',
    avatar varchar(191) NOT NULL,

    email varchar(191) DEFAULT NULL UNIQUE,
    email_verified_at timestamp NULL DEFAULT NULL,

    password varchar(191) DEFAULT NULL UNIQUE,
    phone varchar(191)  DEFAULT NULL,

    CONSTRAINT users_created_at_index UNIQUE (created_at)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists users;
-- +goose StatementEnd
