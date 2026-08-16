// Package store 负责数据库连接的初始化与建表迁移。
//
// 生产环境提示：本项目为了做到"下载即用、零外部依赖"，采用了纯 Go 实现的
// SQLite 驱动（modernc.org/sqlite），SQLite 是单写者模型。若要迁移到生产级
// PostgreSQL，只需要：
//  1. 把这里的 sql.Open("sqlite", ...) 换成 sql.Open("pgx", dsn)
//  2. 把 schema.sql 中的 AUTOINCREMENT 换成 SERIAL / IDENTITY
//  3. 把 service/engine.go 中用于串行化的 sync.Mutex 换成真正的
//     `SELECT ... FOR UPDATE` 行锁（多实例部署时 Mutex 无法跨进程生效）
//
// engine.go 中对这一点有详细注释。
package store

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Open 打开（或创建）SQLite 数据库文件并执行建表语句。
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite 本身是单写者模型，多个 goroutine 并发写会返回 "database is locked"。
	// 将最大打开连接数设为 1，让标准库的连接池天然帮我们排队写请求，
	// 这是 SQLite 场景下最简单可靠的做法。
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("migrate schema: %w", err)
	}

	return db, nil
}
