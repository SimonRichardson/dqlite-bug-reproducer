package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/canonical/go-dqlite/app"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.SetFlags(0)

	fmt.Println("Starting...")

	var sqlite bool
	if len(os.Args) > 1 && os.Args[1] == "sqlite" {
		sqlite = true
	}

	db, close := getDB(sqlite)
	defer close()

	fmt.Println("Running statements")

	// Apply DDL statements.
	if err := tx(context.Background(), db, func(ctx context.Context, tx *sql.Tx) error {
		query := `
CREATE TABLE test (
	id            TEXT PRIMARY KEY,
	value         INT
);
		`
		_, err := tx.ExecContext(ctx, query)
		return err
	}); err != nil {
		log.Fatalf("ERROR: ddl %v", err)
	}

	// Insert watermark
	if err := tx(context.Background(), db, func(ctx context.Context, tx *sql.Tx) error {
		query := `
INSERT INTO test
	(id, value)
VALUES
	(?, -1);
	`
		result, err := tx.ExecContext(ctx, query, "id0")
		if err != nil {
			return err
		}
		_, err = result.RowsAffected()
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatalf("ERROR: insert %v", err)
	}

	// Update watermark
	if err := tx(context.Background(), db, func(ctx context.Context, tx *sql.Tx) error {
		query := `
UPDATE test
SET
	value = ?
WHERE id = ?;
	`
		result, err := tx.ExecContext(ctx, query, 1, "id0")
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected != 1 {
			return fmt.Errorf("expected 1 row affected, got %d", affected)
		}
		return nil
	}); err != nil {
		log.Fatalf("ERROR: update %v", err)
	}

	// Read watermark
	if err := tx(context.Background(), db, func(ctx context.Context, tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, "SELECT value FROM test WHERE id = ?", "id0")
		var value int
		if err := row.Scan(&value); err != nil {
			return err
		}
		if value != 1 {
			return fmt.Errorf("Expected 1 got: %d", value)
		}
		return nil
	}); err != nil {
		log.Fatalf("ERROR: read %v", err)
	}

	fmt.Println("Success")
}

func tx(ctx context.Context, db *sql.DB, fn func(context.Context, *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func getDB(sqlite bool) (*sql.DB, func()) {
	if sqlite {
		fmt.Println("Using sqlite3")

		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			log.Fatalf("ERROR: open %v", err)
		}
		return db, func() { db.Close() }
	}

	fmt.Println("Using dqlite")

	dir, err := ioutil.TempDir("", "dqlite-app-example-")
	if err != nil {
		log.Fatalf("ERROR: tmp dir %v", err)
	}

	node, err := app.New(dir, app.WithAddress("127.0.0.1:9001"))
	if err != nil {
		log.Fatalf("ERROR: new %v", err)
	}

	if err := node.Ready(context.Background()); err != nil {
		log.Fatalf("ERROR: ready %v", err)
	}

	fmt.Println("Ready")

	db, err := node.Open(context.Background(), "test")
	if err != nil {
		log.Fatalf("ERROR: open %v", err)
	}
	return db, func() {
		db.Close()
		node.Close()
		os.RemoveAll(dir)
	}
}
