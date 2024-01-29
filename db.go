// Copyright (C) 2024 Leo Qi <leo@leozqi.com>
//
// Use of this source code is governed by the Apache-2.0 License.
// Full text can be found in the LICENSE file

// Vendors dpapathanasiou/simple graph
// Copyright (c) Denis Papathanasiou
// Used under the terms of the MIT License

package main

import (
    "encoding/json"
    "database/sql"
)

type Webpage struct {
    id string // In this implementation ID doubles as URL
    checksum string // Can change but CRC-32 seems ok, in hex encoding
    keywords []string
    links []string
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
    tx, err := db.Begin()
    if err != nil {
        return err;
    }

    sqlStmt, err := tx.Prepare(`INSERT INTO nodes VALUES(json(?))`)
    if err != nil {
        return err;
    }
    defer sqlStmt.Close()

    _, err = sqlStmt.Exec(json)
    if err != nil {
        return err;
    }

    err = tx.Commit()
    return err
}

func InsertCrawlResult(db *sql.DB, w *Webpage) error {
    b, err := json.Marshal(w);
    if err != nil {
        return err;
    }

    return InsertNode(db, string(b))
}
