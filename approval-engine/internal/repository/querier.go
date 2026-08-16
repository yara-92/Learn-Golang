// Package repository 是数据访问层：只负责 SQL 的拼装与执行，不包含任何业务规则。
//
// 这里的所有函数都接受一个 Querier 接口，而不是直接绑定 *sql.DB。
// *sql.DB 和 *sql.Tx 都实现了这个接口，所以同一套 repository 函数既可以在
// 普通只读查询里用，也可以在 service 层开启的事务里复用——这是 Go 里
// database/sql 最常见的"事务透传"写法，避免为事务单独写一套 repo。
package repository

import (
	"context"
	"database/sql"
)

type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// 确保 *sql.DB 和 *sql.Tx 都满足 Querier（编译期检查）。
var (
	_ Querier = (*sql.DB)(nil)
	_ Querier = (*sql.Tx)(nil)
)
