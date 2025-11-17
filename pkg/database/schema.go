package database

var schemaDefinition = `
create table if not exists picture (
    id integer primary key autoincrement,
    name text not null unique,
    path text not null,
    directory text not null
);

create table if not exists picture_rating (
    id integer primary key autoincrement,
    picture_id integer not null,
    rating integer not null,
    created_at_ts integer not null,

	foreign key (picture_id) references picture(id)
);

create table if not exists picture_stats (
    id integer primary key autoincrement,
    picture_id integer not null unique,
    avg_rating float not null,
    rate_count integer not null,
    
    foreign key (picture_id) references picture(id)
);
`
