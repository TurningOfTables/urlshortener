package main

import (
	"database/sql"

	"github.com/gofiber/fiber/v2/log"
)

func ResetDb(dbPath string) {
	db := ConnectToDb(dbPath)
	log.Infof("Resetting database at %s", dbPath)
	_, err := db.Exec(`DROP TABLE IF EXISTS "links";
		CREATE TABLE "links" (
		"id"	INTEGER UNIQUE,
		"longurl"	TEXT NOT NULL,
		"shorturl"	TEXT NOT NULL,
		"shortcode"	TEXT NOT NULL UNIQUE,
		PRIMARY KEY("id" AUTOINCREMENT)
	);`)
	if err != nil {
		log.Info(err)
		log.Fatal("Database reset failed")
	} else {
		log.Info("Database reset successfully")
	}
}

func ConnectToDb(dbPath string) *sql.DB {
	log.Infof("Connecting to db at %s", dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Info(err)
		log.Fatal("Database connection failed")
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Database ping failed")
	}

	return db
}
