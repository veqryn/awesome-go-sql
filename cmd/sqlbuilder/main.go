/*
Build and run some queries using the sqlbuilder library.
SQLBuilder is focused on flexibly building sql statements, and also has a very
light struct-to-column-names introspection utility, that can be used to avoid
writing out and potentially misspelling or not matching column names in queries.
A very minor disadvantage is that it doesn't use postgres' =ANY($1) format
for IN queries using a slice, and instead enumerates all values in the slice.
It can be used to build dynamic queries, as well as non-standard sql.
It can be combined with other libraries that do struct scanning or querying.
It works with both database/sql and pgx.
*/
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5" // DB Driver
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/veqryn/awesome-go-sql/models"
)

func (d DAO) SelectAccountByID(ctx context.Context, id uint64) (models.AccountIdeal, bool, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	query := sb.Select(
		"id",
		"name",
		"email",
		"active",
		"fav_color",
		"fav_numbers",
		"properties",
		"created_at").
		From("accounts").
		Where(sb.EQ("id", id))

	sqlStr, args := query.Build()

	var account models.AccountIdeal
	err := d.db.QueryRow(ctx, sqlStr, args...).Scan(
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

// sqlbuilder.NewStruct() provides a way to generate the selected columns based
// on a struct and its tags.
var accountModel = sqlbuilder.NewStruct(models.AccountIdeal{}).For(sqlbuilder.PostgreSQL)

func (d DAO) SelectAllAccounts(ctx context.Context) ([]models.AccountIdeal, error) {
	// This generates the selected column names automatically based on the
	// struct type definition.
	query := accountModel.SelectFrom("accounts")

	sqlStr, args := query.Build()

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

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	query := sb.Select(
		"id",
		"name",
		"email",
		"active",
		"fav_color",
		"fav_numbers",
		"properties",
		"created_at").
		From("accounts")

	// Nicely add filters dynamically
	if len(filters.Names) > 0 {
		query = query.Where(sb.In("name", sqlbuilder.List(filters.Names)))
	}
	if filters.Active != nil {
		query = query.Where(sb.EQ("active", *filters.Active))
	}
	if len(filters.FavColors) > 0 {
		query = query.Where(sb.In("fav_color", sqlbuilder.List(filters.FavColors)))
	}

	sqlStr, args := query.Build()
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

	// This is the normal version of pgx
	db, err := pgxpool.New(ctx, "postgresql://postgres:password@localhost:5432/awesome")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dao := DAO{
		db: db,
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
	db *pgxpool.Pool
}
