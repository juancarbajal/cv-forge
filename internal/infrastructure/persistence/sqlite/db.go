package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// OpenDB inicializa la conexión SQLite en la ruta proporcionada y ejecuta las migraciones iniciales
func OpenDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de base de datos: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo base de datos SQLite: %w", err)
	}

	// Configurar PRAGMAs para integridad y concurrencia
	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("error ejecutando pragma '%s': %w", p, err)
		}
	}

	// Ejecutar DDL
	if _, err := db.ExecContext(context.Background(), schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("error ejecutando schema DDL: %w", err)
	}

	return db, nil
}
