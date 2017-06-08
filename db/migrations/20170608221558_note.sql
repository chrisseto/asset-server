-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE note (
  id INTEGER PRIMARY KEY NOT NULL,
  content TEXT NOT NULL,
  created DATETIME NOT NULL,
  modified DATETIME NOT NULL,
  asset_id TEXT NOT NULL,

  FOREIGN KEY(asset_id) REFERENCES asset(id)
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE note
