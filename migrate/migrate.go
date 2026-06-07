package migrate

// ==========================================================
// migrate — 数据库版本迁移管理
//
// 使用 golang-migrate/migrate/v4 管理数据库 schema 变更。
// 替代原来 SQL/ 目录下的手动执行方式。
//
// 使用方法：
//   err := Run(Up, dbType, dsn)
//   err := Run(Down, dbType, dsn)
//
// 迁移文件命名规范：{VERSION}_{DESCRIPTION}.up/down.sql
//   000001_initial_schema.up.sql
//   000001_initial_schema.down.sql
// ==========================================================

import (
	"embed"
	"fmt"
	"strings"

	_ "github.com/glebarez/sqlite"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Direction 迁移方向
type Direction int

const (
	Up   Direction = iota // 升级（应用新迁移）
	Down                   // 回滚（撤销迁移）
)

//go:embed migrations/*.sql
var migrationsEmbed embed.FS

// Run 执行数据库迁移
//   - direction: Up（升级）或 Down（回滚）
//   - dbType: "mysql", "postgres", 或 "sqlite"
//   - dsn: 数据库连接字符串
func Run(direction Direction, dbType string, dsn string) error {
	srcDriver, err := iofs.New(migrationsEmbed, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create source driver: %w", err)
	}

	dbURL := dbTypeToURL(dbType, dsn)
	m, err := migrate.NewWithSourceInstance("iofs", srcDriver, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	switch direction {
	case Up:
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migration up failed: %w", err)
		}
	case Down:
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migration down failed: %w", err)
		}
	}

	version, dirty, _ := m.Version()
	fmt.Printf("Migration complete: version=%d, dirty=%v, direction=%s\n",
		version, dirty, map[Direction]string{Up: "UP", Down: "DOWN"}[direction])
	return nil
}

// MigrateUp 执行所有待应用的迁移（快捷方法）
func MigrateUp(dbType, dsn string) error {
	return Run(Up, dbType, dsn)
}

// MigrateDown 回滚最后一个迁移（快捷方法）
func MigrateDown(dbType, dsn string) error {
	return Run(Down, dbType, dsn)
}

// Version 获取当前数据库版本
func Version(dbType, dsn string) (uint, bool, error) {
	srcDriver, err := iofs.New(migrationsEmbed, "migrations")
	if err != nil {
		return 0, false, err
	}

	dbURL := dbTypeToURL(dbType, dsn)
	m, err := migrate.NewWithSourceInstance("iofs", srcDriver, dbURL)
	if err != nil {
		return 0, false, err
	}
	defer m.Close()

	return m.Version()
}

// dbTypeToURL 构造 migrate 用的 URL
func dbTypeToURL(dbType, dsn string) string {
	switch dbType {
	case "mysql":
		return "mysql://" + dsn
	case "postgres":
		return "postgres://" + dsn
	case "sqlite":
		clean := strings.TrimPrefix(dsn, "sqlite://")
		return "sqlite://" + clean
	default:
		return dsn
	}
}
