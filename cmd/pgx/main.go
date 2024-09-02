/*
Build and run some queries using the pgx package (the latest and most popular
golang postgres driver).
Almost the same as the standard library version, but does not use database/sql,
and therefore has its own interface, though it is almost identical.
It automatically scans into all types that we expect, without additional
wrappers.
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5" // DB Driver
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/veqryn/awesome-go-sql/models"
)

func (d DAO) SelectAccountByID(ctx context.Context, id uint64) (models.AccountIdeal, bool, error) {
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

	var account models.AccountIdeal
	err := d.db.QueryRow(ctx, query, id).Scan(
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

	rows, err := d.db.Query(ctx, query)
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

	rows, err := d.db.Query(ctx, query, args...)
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

	// This is the normal version of pgx
	db, err := pgxpool.New(ctx, "postgresql://postgres:password@localhost:5432/awesome")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dao := DAO{db: db}

	// Query 1
	account, ok, err := dao.SelectAccountByID(ctx, 1)
	if err != nil {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	if !ok {
		panic("ERROR: Account not found")
	}
	fmt.Printf("--------\nQuery by ID\n%s\n", account)

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
	db *pgxpool.Pool
}
