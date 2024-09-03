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
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5" // DB Driver
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/veqryn/awesome-go-sql/models"
)

func (d DAO) SelectAccountByID(ctx context.Context, id uint64) (models.AccountIdeal, bool, error) {
	query := d.builder.Select(
		"id",
		"name",
		"email",
		"active",
		"fav_color",
		"fav_numbers",
		"properties",
		"created_at").
		From("accounts").
		Where(goqu.Ex{"id": id})
	//Prepared(true). // Doesn't work for postgres

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		return models.AccountIdeal{}, false, err
	}

	var account models.AccountIdeal
	err = d.db.QueryRow(ctx, sqlStr, args...).Scan(
		&account.ID,
		&account.Name,
		&account.Email,
		&account.Active,
		&account.FavColor,
		&account.FavNumbers,
		&account.Properties,
		&account.CreatedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return account, false, nil
	case err != nil:
		return account, false, err
	default:
		return account, true, nil
	}
}

func (d DAO) SelectAllAccounts(ctx context.Context) ([]models.AccountIdeal, error) {
	query := d.builder.Select(
		"id",
		"name",
		"email",
		"active",
		"fav_color",
		"fav_numbers",
		"properties",
		"created_at").
		From("accounts")
		//Prepared(true). // Doesn't work for postgres

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.AccountIdeal
	for rows.Next() {
		var account models.AccountIdeal
		scanErr := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Email,
			&account.Active,
			&account.FavColor,
			&account.FavNumbers,
			&account.Properties,
			&account.CreatedAt)
		if scanErr != nil {
			// Check for a scan error. Query rows will be closed with defer.
			return nil, scanErr
		}
		accounts = append(accounts, account)
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (d DAO) SelectAllAccountsByFilter(ctx context.Context, filters models.Filters) ([]models.AccountIdeal, error) {
	query := d.builder.Select(
		"id",
		"name",
		"email",
		"active",
		"fav_color",
		"fav_numbers",
		"properties",
		"created_at").
		From("accounts")
		//Prepared(true). // Doesn't work for postgres

	// Nicely add filters dynamically
	if len(filters.Names) > 0 {
		query = query.Where(goqu.Ex{"name": filters.Names})
	}
	if filters.Active != nil {
		query = query.Where(goqu.Ex{"active": filters.Active})
	}
	if len(filters.FavColors) > 0 {
		query = query.Where(goqu.Ex{"fav_color": filters.FavColors})
	}

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		return nil, err
	}
	fmt.Printf("--------\nDynamic Query SQL:\n%s\n\nDynamic Query Args:\n%#+v\n", sqlStr, args)

	rows, err := d.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.AccountIdeal
	for rows.Next() {
		var account models.AccountIdeal
		scanErr := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Email,
			&account.Active,
			&account.FavColor,
			&account.FavNumbers,
			&account.Properties,
			&account.CreatedAt)
		if scanErr != nil {
			// Check for a scan error. Query rows will be closed with defer.
			return nil, scanErr
		}
		accounts = append(accounts, account)
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func main() {
	ctx := context.Background()

	// This is the database/sql version of pgx
	db, err := pgxpool.New(ctx, "postgresql://postgres:password@localhost:5432/awesome")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dao := DAO{
		builder: goqu.Dialect("postgres"),
		db:      db,
	}

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
	builder goqu.DialectWrapper
	db      *pgxpool.Pool
}
