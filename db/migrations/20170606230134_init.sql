-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE asset (
  id TEXT PRIMARY KEY NOT NULL,
  name VARCHAR(255) NOT NULL,
  deleted BOOLEAN DEFAULT FALSE,
  created DATETIME NOT NULL,
  modified DATETIME NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE asset
