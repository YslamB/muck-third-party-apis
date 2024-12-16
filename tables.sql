drop table if exists apis;

create table apis (
    id serial primary key,
    url character varying(100) not null,
    data text not null,
    method character varying(10) not null,
    status int not null default 200,
    created_at timestamp without time zone not null default now(),

    unique(url, status, method)
);

select 
    json_agg(
        json_build_object(
            'data', data,
            'status', status
        )
    ) results
from apis;
