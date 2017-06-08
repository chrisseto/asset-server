# Asset Server

## Architecture
For this particular service, the focus seemed to be on high availability.
With that in mind, Consistency is being sacrificed.

Rather than individual servers connecting to a database cluster, each server hosts its own database.

The specific database does not matter all that much, this implementation has opted to use SQLite. 
Having a self-hosted database removes another potential point of failure. The server itself would have to crash in order to be come inaccessable.

When starting a server, it may be provided with a list of hostnames to sync with.
On a set interval, 10 seconds in this case, each server will page through the assets of the others, creating assets as need be.
If a conflict is encounted, whichever asset was created first will take precedent.

Currently, notes are not synchronized but would be implemented in the same fashion.
The only additional change would be changing the primary key of notes from an autoincrementing integer to a UUID.

### Alternatively
If consistency was a higher priority, I would opt to switch to a database similar to [cockroachdb](https://github.com/cockroachdb/cockroach/) and mirror its views on CAP Theorem.

## API
> If you’ve ever argued with your team about the way your JSON responses should be formatted, JSON API can be your anti-bikeshedding tool.

The API is [JSON-API](http://jsonapi.org/) compliant, with the root being `/api/v1/`

All resources are ordered by their modification date.

Available resource are `assets` and `notes`

See `scripts/test.py` for API usage examples.

### Avialable Actions (Routes)

* List Notes `GET /api/v1/notes`
* Create a Note `POST /api/v1/notes`
* Read a Note `GET /api/v1/notes/<note-id>`
* Read a Note's Asset `GET /api/v1/notes/<note-id>/asset`
* List Assets `GET /api/v1/assets`
* Create an Asset `POST /api/v1/assets`
* Read an Asset `POST /api/v1/assets/<asset-id>`
* Delete an Asset `DELETE /api/v1/assets/<asset-id>`

Note: Relations for assets are not currently implemented.

### Pagination
Pagination has not been implemented, filtering on the `Modified` field may be used in it's place.

All listing requests are limitted to 100 results at a time.

### Filtering
Filtering is avaiable on any struct field in the form of `filter[FIELD][OP]=value`.

Filter by the content of Name, `filter[Name]=Some%20Name` or `filter[Name][eq]=Some%20Name`.

Filter for anything created after 2017-01-01 , `filter[Created][gt]=2017-01-01`.


## Schema
The Database schema is fairly minimal.

```sql
CREATE TABLE asset (
  id TEXT PRIMARY KEY NOT NULL,
  name VARCHAR(255) NOT NULL,
  deleted BOOLEAN DEFAULT FALSE,
  created DATETIME NOT NULL,
  modified DATETIME NOT NULL
);

CREATE TABLE note (
  id INTEGER PRIMARY KEY NOT NULL,
  content TEXT NOT NULL,
  created DATETIME NOT NULL,
  modified DATETIME NOT NULL,
  asset_id TEXT NOT NULL,

  FOREIGN KEY(asset_id) REFERENCES asset(id)
);
```


## Testing/See it in action

With docker and docker-compose installed, run `docker-compose up leader`.

## Load Testing

Load testing is done using [vegeta](https://github.com/tsenart/vegeta).
To load test locally, fire up the server with `make run` and in another terminal run `make stress`.
Results will be printed to stdout.

See `Dockerfile` for installation instructions.

```
Requests      [total, rate]            1500, 150.10
Duration      [total, attack, wait]    9.996223438s, 9.993332259s, 2.891179ms
Latencies     [mean, 50, 95, 99, max]  2.609128ms, 2.581779ms, 4.35326ms, 5.031091ms, 11.264959ms
Bytes In      [total, mean]            139826, 93.22
Bytes Out     [total, mean]            50660, 33.77
Success       [ratio]                  20.33%
Status Codes  [code:count]             404:15  201:100  204:100  409:394  410:786  200:105
Error Set:
404 Not Found
409 Conflict
410 Gone
```
