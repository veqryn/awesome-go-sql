# awesome-go-sql
Actual example use cases for a curated list of golang sql builder/generator/scanner/helper libraries

## Summary
For static queries, try `sqlc`.

For dynamic queries, try `go-sqlbuilder`, `squirrel`, or `jet` to craft the SQL string and args slice, then use `scany` or `ksql` to run the query and scan to a struct/slice.


## Completed Examples
### No libraries besides database drivers
* [database/sql](./cmd/stdlib/main.go)
* [github.com/jackc/pgx/v5](./cmd/pgx/main.go)

### Model/SQL/Function Generators
* [github.com/sqlc-dev/sqlc](./cmd/sqlc/main.go)
* [github.com/go-jet/jet/v2](./cmd/jet/main.go)

### SQL Builders
* [github.com/huandu/go-sqlbuilder](./cmd/sqlbuilder/main.go)
* [github.com/Masterminds/squirrel](./cmd/squirrel/main.go)
* [github.com/doug-martin/goqu/v9](./cmd/goqu/main.go)

### Struct Scanners 
* [github.com/georgysavva/scany/v2](./cmd/scany/main.go)
* [github.com/vingarcia/ksql](./cmd/ksql/main.go)
* [github.com/blockloop/scan](./cmd/scan/main.go)
* [github.com/jmoiron/sqlx](./cmd/sqlx/main.go)


## Ran Into Problems
* [github.com/bokwoon95/sq](./cmd/sq/main.go) Queries have errors.
* github.com/lqs/sqlingo Failed to generate files for postgres.
