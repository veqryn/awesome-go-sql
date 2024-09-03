/*
Build and run some queries using the goqu library.
GOQU combines both a SQL Builder with a struct scanner.
Because of this, the boilerplate code is cut to an absolute minimum.
Some disadvantages:
  - GOQU does its own interpolation. This is arguably a security concern, as I
    would much rather trust Postgres to correctly interpolate parameters.
  - A very minor disadvantage is that it doesn't use postgres' =ANY($1) format
    for IN queries using a slice, and instead enumerates all values in the slice.
  - The struct scanning only works if you pass the database/sql package's DB
    object in, which means it only works with database/sql and does not work
    with PGX.
    You could choose to only use the builder, and use a different scanning
    library, if you wanted PGX.
*/
package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/jackc/pgx/v5/stdlib" // DB Driver
	"github.com/veqryn/awesome-go-sql/models"
)

func (d DAO) SelectAccountByID(ctx context.Context, id uint64) (models.AccountCompatible, bool, error) {
	var account models.AccountCompatible
	ok, err := d.Select(
		"id",
		"name",
		"email",
		"active",
		"fav_color",
		"fav_numbers",
		"properties",
		"created_at").
		From("accounts").
		Where(goqu.Ex{"id": id}).
		//Prepared(true). // Doesn't work for postgres
		ScanStructContext(ctx, &account)

	return account, ok, err
}

func (d DAO) SelectAllAccounts(ctx context.Context) ([]models.AccountCompatible, error) {
	var accounts []models.AccountCompatible
	err := d.Select(
		"id",
		"name",
		"email",
		"active",
		"fav_color",
		"fav_numbers",
		"properties",
		"created_at").
		From("accounts").
		// Prepared(true). // Doesn't work for postgres
		ScanStructsContext(ctx, &accounts)

	return accounts, err
}

func (d DAO) SelectAllAccountsByFilter(ctx context.Context, filters models.Filters) ([]models.AccountCompatible, error) {
	query := d.Select(
		"id",
		"name",
		"email",
		"active",
		"fav_color",
		"fav_numbers",
		"properties",
		"created_at").
		From("accounts")
		// .Prepared(true) // Doesn't work for postgres

	if len(filters.Names) > 0 {
		query = query.Where(goqu.Ex{"name": filters.Names})
	}
	if filters.Active != nil {
		query = query.Where(goqu.Ex{"active": *filters.Active})
	}
	if len(filters.FavColors) > 0 {
		query = query.Where(goqu.Ex{"fav_color": filters.FavColors})
	}

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		return nil, err
	}

	fmt.Printf("--------\nDynamic Query SQL:\n%s\n\nDynamic Query Args:\n%#+v\n", sqlStr, args)

	var accounts []models.AccountCompatible
	err = d.ScanStructsContext(ctx, &accounts, sqlStr, args...)
	return accounts, err
}

func main() {
	ctx := context.Background()

	// This is the database/sql version of pgx
	db, err := sql.Open("pgx", "postgresql://postgres:password@localhost:5432/awesome")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dao := DAO{Database: goqu.New("postgres", db)}

	// Query 1
	_, ok, err := dao.SelectAccountByID(ctx, 0)
	if err != nil {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	if ok {
		panic("ERROR: Account should not be found")
	}
	// fmt.Printf("--------\nQuery by ID\n%s\n", account)

	// Query multiple
	accounts, err := dao.SelectAllAccounts(ctx)
	if err != nil {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	fmt.Println("--------\nQuery All")
	for _, account := range accounts {
		fmt.Printf("%s\n\n", account)
	}

	// Dynamic Query of multiple
	active := true
	accounts, err = dao.SelectAllAccountsByFilter(ctx, models.Filters{
		Names:     []string{"Jane", "John"},
		Active:    &active,
		FavColors: []string{"red", "blue", "green"},
	})
	if err != nil {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	fmt.Println("--------\nQuery Filter")
	for _, account := range accounts {
		fmt.Printf("%s\n\n", account)
	}
}

type DAO struct {
	*goqu.Database // Wrap the db connection
}
