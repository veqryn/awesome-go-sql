package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Filters exists to test out dynamic querying
type Filters struct {
	Names     []string
	Active    *bool
	FavColors []string
}

// AccountIdeal is the ideal model for an "accounts" row we would like to use,
// with hope that our driver and helper library can directly use this.
type AccountIdeal struct {
	ID         uint64           `json:"id" db:"id" ksql:"id"`
	Name       string           `json:"name" db:"name" ksql:"name"`
	Email      string           `json:"email" db:"email" ksql:"email"`
	Active     bool             `json:"active" db:"active" ksql:"active"`
	FavColor   *string          `json:"fav_color" db:"fav_color" ksql:"fav_color"`       // Can be null // Could also use sql.NullString or Nullable instead of *string
	FavNumbers []int            `json:"fav_numbers" db:"fav_numbers" ksql:"fav_numbers"` // Can be null
	Properties *json.RawMessage `json:"properties" db:"properties" ksql:"properties"`    // Can be null // Could also use []byte instead of *json.RawMessage
	CreatedAt  time.Time        `json:"created_at" db:"created_at" ksql:"created_at"`
}

func (a AccountIdeal) String() string {
	return fmt.Sprintf("Account:\nID: %d\nName: %s\nEmail: %s\nActive: %t\nFavColor: %s\nFavNumbers: %v\nProperties: %s\nCreatedAt: %s",
		a.ID,
		a.Name,
		a.Email,
		a.Active,
		PtrToStr(a.FavColor),
		SliceToStr(a.FavNumbers),
		PtrToStr(a.Properties),
		a.CreatedAt)
}

// AccountCompatible is not as nice as AccountIdeal, because it has to use a
// specialized version of FavNumbers because the helper library can't directly
// scan a postgres array to a golang []int.
// This is usually because the helper library is forced to use the stdlib
// version of pgx.
type AccountCompatible struct {
	ID         uint64           `json:"id" db:"id" ksql:"id"`
	Name       string           `json:"name" db:"name" ksql:"name"`
	Email      string           `json:"email" db:"email" ksql:"email"`
	Active     bool             `json:"active" db:"active" ksql:"active"`
	FavColor   *string          `json:"fav_color" db:"fav_color" ksql:"fav_color"`       // Can be null // Could also use sql.NullString or Nullable instead of *string
	FavNumbers Array[int]       `json:"fav_numbers" db:"fav_numbers" ksql:"fav_numbers"` // Can be null
	Properties *json.RawMessage `json:"properties" db:"properties" ksql:"properties"`    // Can be null // Could also use []byte instead of *json.RawMessage
	CreatedAt  time.Time        `json:"created_at" db:"created_at" ksql:"created_at"`
}

func (a AccountCompatible) String() string {
	return fmt.Sprintf("Account:\nID: %d\nName: %s\nEmail: %s\nActive: %t\nFavColor: %s\nFavNumbers: %v\nProperties: %s\nCreatedAt: %s",
		a.ID,
		a.Name,
		a.Email,
		a.Active,
		PtrToStr(a.FavColor),
		SliceToStr(a.FavNumbers),
		PtrToStr(a.Properties),
		a.CreatedAt)
}

// Array is a wrapper that allows scanning a postgres Array into a golang slice
type Array[T any] []T

func (a *Array[T]) Scan(src any) error {
	var dst []T
	if err := pgTypeMap.SQLScanner(&dst).Scan(src); err != nil {
		return err
	}
	*a = Array[T](dst)
	return nil
}

func (a Array[T]) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	src := pgtype.FlatArray[T](a)
	arrayType, ok1 := pgTypeMap.TypeForValue(src)
	elementType, ok2 := pgTypeMap.TypeForValue(src[0])
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("can't encode %T", src)
	}
	codec := pgtype.ArrayCodec{ElementType: elementType}

	buf, err := codec.PlanEncode(pgTypeMap, arrayType.OID, pgtype.TextFormatCode, src).Encode(src, nil)
	if err != nil {
		return nil, err
	}
	return string(buf), err
}

func (a Array[T]) Get() []T {
	return a
}

var pgTypeMap = pgtype.NewMap()

/*
The following are just helpers for our own debugging and pretty-printing
*/

// PtrToStr lets us print pointers for debug purposes
func PtrToStr[T any](v *T) string {
	if v == nil {
		return "<nil>"
	}
	// Hack for displaying []byte as a string
	if refType := reflect.TypeOf(*v); refType.Kind() == reflect.Slice && refType.Elem().Kind() == reflect.Uint8 {
		return fmt.Sprintf("%s", *v)
	}
	return fmt.Sprint(*v)
}

// SliceToStr lets us print slices for debug purposes, and know if they are nil or not
func SliceToStr[T any](v []T) string {
	if v == nil {
		return "<nil>"
	}
	// Hack for displaying []byte as a string
	if refType := reflect.TypeOf(v); refType.Kind() == reflect.Slice && refType.Elem().Kind() == reflect.Uint8 {
		return fmt.Sprintf("%s", v)
	}
	return fmt.Sprint(v)
}

// Nullable wraps sql.Null to provide a String method
type Nullable[T any] struct {
	sql.Null[T]
}

func (n Nullable[T]) String() string {
	if !n.Valid {
		return "<nil>"
	}

	// Hack for displaying []byte as a string
	if refType := reflect.TypeOf(n.V); refType.Kind() == reflect.Slice && refType.Elem().Kind() == reflect.Uint8 {
		return fmt.Sprintf("%s", n.V)
	}
	return fmt.Sprintf("%v", n.V)
}
