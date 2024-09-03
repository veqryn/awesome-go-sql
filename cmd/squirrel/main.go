/*
Build and run some queries using the Squirrel library.
The main advantage of this library is using it to build dynamic queries,
though it can be used to build any queries.
A very minor disadvantage is that it doesn't use postgres' =ANY($1) format
for IN queries using a slice, and instead enumerates all values in the slice.
Because it only creates the SQL string and argument list, it can be combined
with other libraries that do struct scanning or querying.
It works with both database/sql and pgx.
*/
package main

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5" // DB Driver
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/veqryn/awesome-go-sql/models"
)

func (d DAO) SelectAccountByID(ctx context.Context, id uint64) (models.AccountIdeal, bool, error) {
	query := sq.
		Select(
			"id",
			"name",
			"email",
			"active",
			"fav_color",
			"fav_numbers",
			"properties",
			"created_at").
		From("accounts").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
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
	query := sq.
		Select(
			"id",
			"name",
			"email",
			"active",
			"fav_color",
			"fav_numbers",
			"properties",
			"created_at").
		From("accounts").
		OrderBy("id")

	sqlStr, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
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
	query := sq.
		Select(
			"id",
			"name",
			"email",
			"active",
			"fav_color",
			"fav_numbers",
			"properties",
			"created_at").
		From("accounts").
		OrderBy("id")

	// Nicely add filters dynamically
	if len(filters.Names) > 0 {
		query = query.Where(sq.Eq{"name": filters.Names})
	}
	if filters.Active != nil {
		query = query.Where(sq.Eq{"active": *filters.Active})
	}
	if len(filters.FavColors) > 0 {
		query = query.Where(sq.Eq{"fav_color": filters.FavColors})
	}

	sqlStr, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
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

	// This is the normal version of pgx
	db, err := pgxpool.New(ctx, "postgresql://postgres:password@localhost:5432/awesome")
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
	db *pgxpool.Pool
}
