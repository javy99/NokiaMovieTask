package main

import (
	"database/sql"
	"log"
	"testing"
	_ "github.com/go-sql-driver/mysql"
)

func connectTestDB() *sql.DB {
	dsn := "root:javyroot@tcp(localhost:3306)/movie_db"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot reach the database: %v", err)
	}
	return db
}

func TestConnectDB(t *testing.T) {
	db := connectTestDB()
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
}

func TestAddPerson(t *testing.T) {
	db := connectTestDB()
	defer db.Close()

	AddPerson(db, "Test Person", 1980)

	var id int
	err := db.QueryRow("SELECT id FROM people WHERE name = ?", "Test Person").Scan(&id)
	if err != nil {
		t.Fatalf("Failed to find added person: %v", err)
	}

	_, err = db.Exec("DELETE FROM people WHERE id = ?", id)
	if err != nil {
		t.Fatalf("Failed to delete test person: %v", err)
	}
}

func TestAddMovie(t *testing.T) {
	db := connectTestDB()
	defer db.Close()

	AddPerson(db, "Test Director", 1970)
	AddPerson(db, "Test Actor 1", 1985)
	AddPerson(db, "Test Actor 2", 1990)

	actors := []string{"Test Actor 1", "Test Actor 2"}
	AddMovie(db, "Test Movie", 120, "Test Director", 2020, actors)

	var movieID int
	err := db.QueryRow("SELECT id FROM movies WHERE title = ? AND director_id = (SELECT id FROM people WHERE name = ?)", "Test Movie", "Test Director").Scan(&movieID)
	if err != nil {
		t.Fatalf("Failed to find added movie: %v", err)
	}

	_, err = db.Exec("DELETE FROM movie_actors WHERE movie_id = ?", movieID)
	if err != nil {
	    t.Fatalf("Failed to delete actors from movie_actors table: %v", err)
	}
	
	_, err = db.Exec("DELETE FROM movies WHERE id = ?", movieID)
	if err != nil {
	    t.Fatalf("Failed to delete test movie: %v", err)
	}
	
	_, err = db.Exec("DELETE FROM people WHERE name IN ('Test Director', 'Test Actor 1', 'Test Actor 2')")
	if err != nil {
	    t.Fatalf("Failed to delete test persons: %v", err)
}
}

func TestDeletePerson(t *testing.T) {
	db := connectTestDB()
	defer db.Close()

	AddPerson(db, "Test Delete Person", 1990)

	DeletePerson(db, "Test Delete Person")

	var id int
	err := db.QueryRow("SELECT id FROM people WHERE name = ?", "Test Delete Person").Scan(&id)
	if err == nil {
		t.Fatalf("Person was not deleted")
	}
}
