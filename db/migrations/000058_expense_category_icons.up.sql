-- Give expense categories a visual identity (icon + accent colour) so the
-- admin UI can show them as an icon list rather than plain text, and seed
-- the common ones every shop has. icon holds an mdi glyph name
-- ("mdi-flash"); color holds a Vuetify colour token ("amber").
alter table expense_categories add column if not exists icon text not null default 'mdi-tag-outline';
alter table expense_categories add column if not exists color text not null default 'blue-grey';

-- Predefined categories. On a fresh install these are inserted outright; on
-- an install where 000041 already backfilled the name from real data, the
-- ON CONFLICT branch fills in the icon/colour *only if it's still the
-- freshly-added default* -- so an admin who later customises a category's
-- icon is never overwritten by a re-run, but existing installs still get
-- the nice defaults.
insert into expense_categories (name, icon, color, is_active) values
  ('Electricity',  'mdi-flash',                   'amber',           true),
  ('Water',        'mdi-water',                   'blue',            true),
  ('Rent',         'mdi-home-city',               'deep-purple',     true),
  ('Salary',       'mdi-cash-multiple',           'green',           true),
  ('Internet',     'mdi-wifi',                    'cyan',            true),
  ('Phone',        'mdi-phone',                   'teal',            true),
  ('Maintenance',  'mdi-wrench',                  'blue-grey',       true),
  ('Supplies',     'mdi-package-variant-closed',  'orange',          true),
  ('Transport',    'mdi-truck',                   'indigo',          true),
  ('Fuel',         'mdi-gas-station',             'deep-orange',     true),
  ('Marketing',    'mdi-bullhorn',                'pink',            true),
  ('Taxes',        'mdi-bank',                    'red',             true),
  ('Bank Fees',    'mdi-credit-card-outline',     'brown',           true),
  ('Cleaning',     'mdi-broom',                   'light-green',     true),
  ('Insurance',    'mdi-shield-check',            'blue-darken-2',   true),
  ('Meals',        'mdi-food',                    'orange-darken-2', true)
on conflict (name) do update
  set icon = excluded.icon, color = excluded.color
  where expense_categories.icon = 'mdi-tag-outline'
    and expense_categories.color = 'blue-grey';
