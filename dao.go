package main

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type dao struct {
	connPool *pgxpool.Pool
}

func NewDao() *dao {
	return &dao{}
}

func (dao *dao) Connect() error {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	var err error
	dao.connPool, err = pgxpool.Connect(
		context.Background(),
		os.Getenv("DATABASE_URL"),
	)
	return err
}

func (dao *dao) Close() {
	dao.connPool.Close()
}

func (dao *dao) Query(query string) (*pgx.Rows, error) {
	rows, err := dao.connPool.Query(context.Background(), query)
	if err != nil {
		panic(fmt.Errorf("QueryRow failed: %v\n", err))
	}
	return &rows, err
}

func (dao *dao) Scan(query string, resultSet []interface{}) {
	// filling provided resultset structure \with query result
	pgxscan.Select(context.Background(), dao.connPool, &resultSet, query)
}

func (dao *dao) Iterate(rows *pgx.Rows, targetType interface{}) interface{} {
	defer (*rows).Close()

	// get target type - reflect.Type
	t := reflect.TypeOf(targetType)

	//var result []interface{}
	result := reflect.Zero(reflect.SliceOf(t))
	for (*rows).Next() {
		// create new instance of type t
		instance := reflect.New(t)

		// scan to instance
		(*rows).Scan(&instance)

		// append to result
		reflect.Append(result, instance)
	}
	return result.Interface()
}
