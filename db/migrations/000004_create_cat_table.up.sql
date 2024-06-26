CREATE TABLE IF NOT EXISTS cat(
        id serial PRIMARY KEY,
        user_id BIGINT NOT NULL,

        name text not null,
        image_urls text[] not null,
        sex cat_sex not null,
        race cat_race not null,
        age int not null default 1,
        description text not null,
        has_matched boolean not null default false,

        created_at timestamptz NOT NULL DEFAULT now(),
        updated_at timestamptz NOT NULL DEFAULT now(),
        deleted_at timestamptz DEFAULT NULL,

        CONSTRAINT fk_users_cats foreign key(user_id) references public.users(id)
);

create index if not exists idx_cat_id on cat(id); 
create index if not exists idx_cat_user_id on cat(user_id); 