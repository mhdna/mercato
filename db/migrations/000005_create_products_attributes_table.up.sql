CREATE TABLE IF NOT EXISTS attributes (
  id bigint primary key,
  name text not null
);

CREATE TABLE IF NOT EXISTS attributes_values (
  -- we MUST have an ID here since we might have many values per attribute
  -- later on, we might need to update the value for example
  id bigserial primary key,
  attribute_id bigint not null references attributes(id),
  value text not null,
  UNIQUE (attribute_id, value)
);

CREATE TABLE IF NOT EXISTS products_attributes (
  attribute_id bigint not null references attributes(id),
  attribute_value_id bigint not null references attributes_values(id),
  product_id bigint not null references products(id),
  PRIMARY KEY (attribute_id, product_id)
);

INSERT INTO attributes
(name, id) VALUES
  ('category',0),
  ('sub-category',1),
  ('brand',2),
  ('kind',3),
  ('type',4),
  ('unit',5),
  ('year',6),
  ('season',7),
  ('origin',8);
