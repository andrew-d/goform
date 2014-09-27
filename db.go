package main

import ()

type Respondent struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Token string `db:"token"`
}

type Response struct {
	ID         int64  `db:"id"`
	Respondent int64  `db:"respondent"`
	Item       string `db:"item"`
	Quantity   int    `db:"quantity"`
	MaxPrice   int    `db:"max_price"`
	Notes      string `db:"notes"`
}

var createStatements = []string{
	`CREATE TABLE IF NOT EXISTS respondents (
		id     INTEGER PRIMARY KEY AUTOINCREMENT,
		name   TEXT NOT NULL,
		token  TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS responses (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		respondent INTEGER NOT NULL,
		item       TEXT NOT NULL,
		quantity   INTEGER NOT NULL,
		max_price  INTEGER NOT NULL,
		notes      TEXT NOT NULL,

		FOREIGN KEY(respondent) REFERENCES respondents(id)
	)`,
}
