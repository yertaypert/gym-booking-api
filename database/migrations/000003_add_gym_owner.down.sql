-- down migration for 000003
-- This is tricky because we can't easily remove an enum value in postgres.
-- Usually we just don't down migrate enums unless necessary.
-- But we can remove the column.
ALTER TABLE gyms DROP COLUMN owner_id;
