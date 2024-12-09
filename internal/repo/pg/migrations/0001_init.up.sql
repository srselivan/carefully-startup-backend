create table if not exists backend.settings
(
    id                     serial primary key,
    rounds_count           integer not null check (rounds_count >= 3),
    rounds_duration        bigint  not null,
    link_to_pdf            text,
    enable_random_events   boolean,
    default_balance_amount bigint  not null
);

insert into backend.settings (rounds_count, rounds_duration, link_to_pdf, enable_random_events, default_balance_amount)
values (3, 600000000000, 'http://hui/pizda', false, 1000);

create table if not exists backend.game
(
    id            serial primary key,
    state         smallint not null,
    current_round integer,
    trade_state   smallint not null,
    current_game  bigint not null
);

insert into backend.game (state, current_round, trade_state, current_game) values (3, 0, 0, 1);

create table if not exists backend.company
(
    id       bigserial primary key,
    name     text not null,
    archived boolean
);

create table if not exists backend.company_share
(
    id         bigserial primary key,
    company_id bigint  not null references backend.company (id),
    round      integer not null,
    price      bigint  not null,
    unique (company_id, round)
);

create table if not exists backend.additional_info
(
    id          bigserial primary key,
    name        text     not null,
    description text     not null,
    type        smallint not null,
    cost        bigint   not null,
    company_id  bigint references backend.company (id)
);

create table if not exists backend.random_event
(
    id   bigserial primary key,
    name text
);

create table if not exists backend.balance
(
    id     bigserial primary key,
    amount bigint not null
);

create table if not exists backend.team
(
    id                  bigserial primary key,
    created_at          timestamptz not null default now(),
    updated_at          timestamptz,
    name                text        not null,
    members             text[],
    credentials         text        not null unique,
    balance_id          bigint      not null references backend.balance (id),
    shares              jsonb,
    additional_info_ids jsonb,
    random_event_id     bigint references backend.random_event (id),
    game_id             bigint      not null
);

create table if not exists backend.balance_transaction
(
    id                 bigserial primary key,
    balance_id         bigint  not null references backend.balance (id),
    round              integer not null,
    amount             bigint  not null,
    details            jsonb,
    additional_info_id bigint references backend.additional_info (id),
    random_event_id    bigint references backend.random_event (id)
);

create table if not exists backend.team_refresh_token
(
    team_id bigint not null unique,
    token   text   not null
);