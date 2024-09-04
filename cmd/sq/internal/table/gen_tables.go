package table

import "github.com/bokwoon95/sq"

type ACCOUNTS struct {
	sq.TableStruct
	ID          sq.NumberField  `ddl:"type=bigint notnull primarykey default=nextval('accounts_id_seq'::regclass)"`
	NAME        sq.StringField  `ddl:"type=varchar(50) notnull"`
	EMAIL       sq.StringField  `ddl:"type=varchar(50) notnull unique"`
	ACTIVE      sq.BooleanField `ddl:"type=boolean notnull"`
	FAV_COLOR   sq.EnumField    `ddl:"type=colors"`
	FAV_NUMBERS sq.ArrayField   `ddl:"type=int[]"`
	PROPERTIES  sq.JSONField    `ddl:"type=jsonb"`
	CREATED_AT  sq.TimeField    `ddl:"type=timestamptz notnull"`
}
