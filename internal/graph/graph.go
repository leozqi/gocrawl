// Copyright (C) 2024 Leo Qi <leo@leozqi.com>

// dpapathanasiou/simple graph
// Copyright (c) Denis Papathanasiou
// Used under terms of the MIT License

package graph

import (
    "encoding/json"
    "database/sql"
    "errors"
)

type Webpage struct {
    Id string           `json:"id"`       // In this implementation ID doubles as URL
    Checksum string     `json:"checksum"` // Can change but CRC-32 seems ok, in hex encoding
    Keywords []string   `json:"keywords"`
    Links []string      `json:"links"`
}


// Why use JSON?
// Specced textual format supported by many other programs that may acces the
// database
// When reading and writing to the database, we read the whole file and it
// becomes a native go struct.
func InitDbFile(db *sql.DB) error {
    sqlStmt := `
        CREATE TABLE IF NOT EXISTS nodes (
            body TEXT,
            id   TEXT GENERATED ALWAYS AS (json_extract(body, '$.id')) VIRTUAL NOT NULL UNIQUE
        );

        CREATE INDEX IF NOT EXISTS id_idx ON nodes(id);

        CREATE TABLE IF NOT EXISTS edges (
            source     TEXT,
            target     TEXT,
            properties TEXT,
            UNIQUE(source, target, properties) ON CONFLICT REPLACE,
            FOREIGN KEY(source) REFERENCES nodes(id),
            FOREIGN KEY(target) REFERENCES nodes(id)
        );

        CREATE VIRTUAL TABLE IF NOT EXISTS nodes_fts USING fts5(
            id,
            body,
            content=nodes
        );

        CREATE INDEX IF NOT EXISTS source_idx ON edges(source);
        CREATE INDEX IF NOT EXISTS target_idx ON edges(target);
    `

    _, err := db.Exec(sqlStmt)
    return err
}

func InsertNode(db *sql.DB, json string) error {
    // Check string
    if !(len(json) > 0) {
        return errors.New("Empty JSON passed to InsertNode")
    }

    tx, err := db.Begin()
    if err != nil {
        return err
    }

    sqlStmt, err := tx.Prepare(`INSERT INTO nodes VALUES(json(?))`)
    if err != nil {
        return err
    }
    defer sqlStmt.Close()

    _, err = sqlStmt.Exec(json)
    if err != nil {
        return err
    }

    err = tx.Commit()
    return err
}

func InsertCrawlResult(db *sql.DB, w *Webpage) error {
    b, err := json.Marshal(w)
    if err != nil {
        return err
    }

    return InsertNode(db, string(b))
}
