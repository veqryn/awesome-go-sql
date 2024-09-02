/*
Build and run some queries using the SQLX library.
The main advantage of this library is automatic scanning of rows into structs,
or slices of structs.
This cuts down on the boilerplate considerably.
However, SQLX embeds the database/sql package's DB object, meaning that it does
not work directly with PGX, and it also provides a very large api surface area.
Because SQLX is only concerned with querying and scanning, it can be combined
with a query builder library that provides the actual query strings.
*/
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // DB Driver
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
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
	err := d.db.GetContext(ctx, &account, query, id)
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
	err := d.db.Select(&accounts, query)
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
	err := d.db.Select(&accounts, query, args...)
	return accounts, err
}

func main() {
	ctx := context.Background()

	// This is the database/sql version of pgx
	db, err := sqlx.Connect("pgx", "postgresql://postgres:password@localhost:5432/awesome")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// You can override which tags are used for scanning to a struct
	db.Mapper = reflectx.NewMapperFunc("json", nil)

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
	db *sqlx.DB
}
