drop table if exists apis;
create table apis (
    id serial primary key,
    url character varying(100) not null,
    data text not null,
    created_at timestamp without time zone not null default now()
);

