CREATE TABLE IF NOT EXISTS products (
  id bigserial PRIMARY KEY,
  code text not null UNIQUE,
  name text NOT NULL,
  description text not null,
  is_active bool not null DEFAULT true,
  created_at timestamp(0) WITH time zone NOT NULL DEFAULT NOW()
);
