/*
Build and run some queries using the scan library.
Scan is similar to SQLX in that it just scans the results into structs/slices.
Unlike SQLX, it operates on the 'rows' returned by the query, and therefore
it has a much smaller api surface area.
Because scan is only concerned with querying and scanning, it can be combined
with a query builder library that provides the actual query strings.
Unfortunately, scan does not work with PGX.
*/
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/blockloop/scan/v2"
	_ "github.com/jackc/pgx/v5/stdlib" // DB Driver
	"github.com/veqryn/awesome-go-sql/models"
)

func (d DAO) SelectAccountByID(ctx context.Context, id uint64) (models.AccountCompatible, bool, error) {
	const query = `
		SELECT
			id,
			name,
			email,
			active,
			fav_color,
			fav_numbers,
			properties,
			created_at
		FROM accounts
		WHERE id = $1`

	var account models.AccountCompatible
	rows, err := d.db.QueryContext(ctx, query, id)
	if err != nil {
		return models.AccountCompatible{}, false, err
	}

	err = scan.Row(&account, rows)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return account, false, nil
	case err != nil:
		return account, false, err
	default:
		return account, true, nil
	}
}

func (d DAO) SelectAllAccounts(ctx context.Context) ([]models.AccountCompatible, error) {
	const query = `
		SELECT
			id,
			name,
			email,
			active,
			fav_color,
			fav_numbers,
			properties,
			created_at
		FROM accounts
		ORDER BY id`

	var accounts []models.AccountCompatible
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	err = scan.Rows(&accounts, rows)
	return accounts, err
}

func (d DAO) SelectAllAccountsByFilter(ctx context.Context, filters models.Filters) ([]models.AccountCompatible, error) {
	query := `
		SELECT
			id,
			name,
			email,
			active,
			fav_color,
			fav_numbers,
			properties,
			created_at
		FROM accounts`

	// Sadly, we have to manually build dynamic queries
	var wheres []string
	var args []any
	argCount := 1
	if len(filters.Names) > 0 {
		wheres = append(wheres, fmt.Sprintf("name = ANY($%d)", argCount))
		args = append(args, filters.Names)
		argCount++
	}
	if filters.Active != nil {
		wheres = append(wheres, fmt.Sprintf("active = $%d", argCount))
		args = append(args, *filters.Active)
		argCount++
	}
	if len(filters.FavColors) > 0 {
		wheres = append(wheres, fmt.Sprintf("fav_color = ANY($%d)", argCount))
		args = append(args, filters.FavColors)
		argCount++
	}

	if len(wheres) > 0 {
		query += " WHERE " + strings.Join(wheres, " AND ")
	}
	fmt.Printf("--------\nDynamic Query SQL:\n%s\n\nDynamic Query Args:\n%#+v\n", query, args)

	var accounts []models.AccountCompatible
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	err = scan.Rows(&accounts, rows)
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

	dao := DAO{db: db}

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
	db *sql.DB
}
