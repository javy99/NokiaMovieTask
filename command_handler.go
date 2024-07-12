package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

func ListActorsForMovie(db *sql.DB, title, director string) {
	query := `
		SELECT p.name, p.birth_year, m.release_year
		FROM people p
		JOIN movie_actors ma ON p.id = ma.actor_id
		JOIN movies m ON ma.movie_id = m.id
		WHERE m.title = ? AND m.director_id = (SELECT id FROM people WHERE name = ?)
	`
	rows, err := db.Query(query, title, director)
	if err != nil {
		log.Fatalf("Error fetching actors: %v", err)
	}
	defer rows.Close()

	fmt.Println("Starring:")
	for rows.Next() {
		var name string
		var birthYear, releaseYear int
		if err := rows.Scan(&name, &birthYear, &releaseYear); err != nil {
			log.Fatalf("Error scanning actor: %v", err)
		}
		age := releaseYear - birthYear
		fmt.Printf("  - %s at age %d\n", name, age)
	}
}

func ListMovies(db *sql.DB, verbose bool, titleFilter, directorFilter, actorFilter string, orderByLengthAsc, orderByLengthDesc bool) {
	query := `
		SELECT m.title, p.name, m.release_year, m.length
		FROM movies m
		JOIN people p ON m.director_id = p.id
	`
	conditions := []string{}
	params := []interface{}{}

	if titleFilter != "" {
		conditions = append(conditions, "m.title REGEXP ?")
		params = append(params, titleFilter)
	}
	if directorFilter != "" {
		conditions = append(conditions, "p.name REGEXP ?")
		params = append(params, directorFilter)
	}
	if actorFilter != "" {
		conditions = append(conditions, "m.id IN (SELECT ma.movie_id FROM movie_actors ma JOIN people a ON ma.actor_id = a.id WHERE a.name REGEXP ?)")
		params = append(params, actorFilter)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if orderByLengthAsc {
		query += " ORDER BY m.length ASC, m.title ASC"
	} else if orderByLengthDesc {
		query += " ORDER BY m.length DESC, m.title ASC"
	} else {
		query += " ORDER BY m.title"
	}

	rows, err := db.Query(query, params...)
	if err != nil {
		log.Fatalf("Error fetching movies: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var title, director string
		var year, length int
		if err := rows.Scan(&title, &director, &year, &length); err != nil {
			log.Fatalf("Error scanning movie: %v", err)
		}

		lengthFormatted := fmt.Sprintf("%02d:%02d", length/60, length%60)
		if verbose {
			fmt.Printf("%s by %s in %d, %s\n", title, director, year, lengthFormatted)
			ListActorsForMovie(db, title, director)
		} else {
			fmt.Printf("%s by %s in %d, %s\n", title, director, year, lengthFormatted)
		}
	}
}

func AddPerson(db *sql.DB, name string, birthYear int) {
	query := `
		INSERT INTO people (name, birth_year)
		VALUES (?, ?)
	`
	_, err := db.Exec(query, name, birthYear)
	if err != nil {
		log.Fatalf("Error adding person: %v", err)
	}
	fmt.Println("Person added successfully!")
}

func AddMovie(db *sql.DB, title string, length int, director string, releaseYear int, actors []string) {
	var existingMovieID int
	err := db.QueryRow("SELECT id FROM movies WHERE title = ? AND director_id = (SELECT id FROM people WHERE name = ?)", title, director).Scan(&existingMovieID)
	if err == nil {
		fmt.Println("Movie with the same title and director already exists!")
		return
	} else if err != sql.ErrNoRows {
		log.Fatalf("Error checking existing movie: %v", err)
	}

	var directorID int
	err = db.QueryRow("SELECT id FROM people WHERE name = ?", director).Scan(&directorID)
	if err != nil {
		log.Fatalf("Director not found: %v", err)
	}

	query := `
		INSERT INTO movies (title, length, director_id, release_year)
		VALUES (?, ?, ?, ?)
	`
	res, err := db.Exec(query, title, length, directorID, releaseYear)
	if err != nil {
		log.Fatalf("Error adding movie: %v", err)
	}

	movieID, err := res.LastInsertId()
	if err != nil {
		log.Fatalf("Error getting last insert ID: %v", err)
	}

	for _, actor := range actors {
		var actorID int
		err := db.QueryRow("SELECT id FROM people WHERE name = ?", actor).Scan(&actorID)
		if err != nil {
			log.Fatalf("Actor not found: %v", err)
		}

		query := `
			INSERT INTO movie_actors (movie_id, actor_id)
			VALUES (?, ?)
		`
		_, err = db.Exec(query, movieID, actorID)
		if err != nil {
			log.Fatalf("Error adding actor to movie: %v", err)
		}
	}
	fmt.Println("Movie added successfully!")
}

func DeletePerson(db *sql.DB, name string) {
	var id int
	err := db.QueryRow("SELECT id FROM people WHERE name = ?", name).Scan(&id)
	if err != nil {
		fmt.Println("Person not found.")
		return
	}

	var directorInMovie bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM movies WHERE director_id = ?)", id).Scan(&directorInMovie)
	if err != nil {
		log.Fatalf("Error checking if person is a director: %v", err)
	}

	if directorInMovie {
		fmt.Println("Cannot delete person. They are a director of a movie.")
		return
	}

	_, err = db.Exec("DELETE FROM movie_actors WHERE actor_id = ?", id)
	if err != nil {
		log.Fatalf("Error deleting person from movie_actors: %v", err)
	}

	_, err = db.Exec("DELETE FROM people WHERE id = ?", id)
	if err != nil {
		log.Fatalf("Error deleting person: %v", err)
	}

	fmt.Println("Person deleted successfully!")
}