package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
    dsn := "root:javyroot@tcp(localhost:3306)/movie_db"
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Error connecting to the database: %v", err)
    }
    defer db.Close()

    error := db.Ping();
	if error != nil {
        log.Fatalf("Cannot reach the database: %v", err)
    }

    fmt.Println("Successfully connected to the database!")

    handleCommands(db)
}


func handleCommands(db *sql.DB) {
	fmt.Println("============================================================================")
	fmt.Println("-------------- Welcome to the NokiaMovie Console Application! --------------")
	printCommandDocumentation()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				log.Fatalf("Error reading standard input: %v", err)
			}
			break
		}

		input := scanner.Text()
		parts := strings.Fields(input) // Split by spaces and remove extra spaces
		
		if len(parts) == 0 {
			fmt.Println("Please enter a command.")
			continue
		}

		// l -v
		command := parts[0] // l
		args := parts[1:]   // [-v]

		switch command {
		case "l":
			handleListMovies(db, args)
		case "a":
			handleAddCommands(db, args, scanner)
		case "d":
			handleDeleteCommands(db, args, scanner)
		case "h":
			printCommandDocumentation()
		case "exit":
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Invalid command. Type 'h' for help.")
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading standard input: %v", err)
	}
}

func handleListMovies(db *sql.DB, args []string) {
	var verbose, orderByLengthAsc, orderByLengthDesc bool
	var titleFilter, directorFilter, actorFilter string

	// Combine args into a single string to handle quotes
	input := strings.Join(args, " ")
	fmt.Println(input)
	
	// Define regex to match flags and quoted arguments
	re := regexp.MustCompile(`-(\w)(?:\s+"([^"]+)"|\s+(\S+))?`)
	fmt.Println(re)

	// Find all matches
	matches := re.FindAllStringSubmatch(input, -1)
	fmt.Println(matches)
	for _, match := range matches {
		fmt.Println(match)

		flag := match[1]
		fmt.Println(flag)
		var value string
		if match[2] != "" {
			value = match[2]
		} else if match[3] != "" {
			value = match[3]
		}

		switch flag {
		case "v":
			verbose = true
		case "t":
			if value != "" {
				titleFilter = value // Star

			} else {
				fmt.Println("No regex provided for -t switch")
				return
			}
		case "d":
			if value != "" {
				directorFilter = value
			} else {
				fmt.Println("No regex provided for -d switch")
				return
			}
		case "a":
			if value != "" {
				actorFilter = value
			} else {
				fmt.Println("No regex provided for -a switch")
				return
			}
		case "la":
			if orderByLengthDesc {
				fmt.Println("Cannot use both -la and -ld switches")
				return
			}
			orderByLengthAsc = true
		case "ld":
			if orderByLengthAsc {
				fmt.Println("Cannot use both -la and -ld switches")
				return
			}
			orderByLengthDesc = true
		default:
			fmt.Printf("Unknown switch: %s\n", flag)
			return
		}
	}

	if orderByLengthAsc && orderByLengthDesc {
		fmt.Println("Invalid query: both -la and -ld cannot be specified simultaneously.")
		return
	}

	listMovies(db, verbose, titleFilter, directorFilter, actorFilter, orderByLengthAsc, orderByLengthDesc)
}

func handleAddCommands(db *sql.DB, args []string, scanner *bufio.Scanner) {
	if len(args) > 0 {
		subCommand := args[0]
		switch subCommand {
		case "-p":
			addPersonInteractive(db, scanner)
		case "-m":
			addMovieInteractive(db, scanner)
		default:
			fmt.Println("Unknown sub-command for add")
		}
	} else {
		fmt.Println("No sub-command provided for add")
	}
}

func addMovieInteractive(db *sql.DB, scanner *bufio.Scanner) {
	var title, director string
	var length string
	var releaseYear int
	var actors []string
	var hours, minutes int

	fmt.Print("Title: ")
	if !scanner.Scan() {
		log.Fatalf("Error reading title: %v", scanner.Err())
	}
	title = scanner.Text()

	for {
		fmt.Print("Length (hh:mm): ")
		if !scanner.Scan() {
			log.Fatalf("Error reading length: %v", scanner.Err())
		}
		length = scanner.Text()

		// Parse length into hours and minutes
		if _, err := fmt.Sscanf(length, "%d:%d", &hours, &minutes); err != nil || hours < 0 || minutes < 0 || minutes >= 60 {
			fmt.Println("Bad input format (hh:mm), try again!")
		} else {
			break
		}
	}

	for {
		fmt.Print("Director: ")
		if !scanner.Scan() {
			log.Fatalf("Error reading director: %v", scanner.Err())
		}
		director = scanner.Text()

		var directorID int
		err := db.QueryRow("SELECT id FROM people WHERE name = ?", director).Scan(&directorID)
		if err != nil {
			fmt.Printf("- We could not find \"%s\", try again!\n", director)
		} else {
			break
		}
	}

	fmt.Print("Release Year: ")
	if _, err := fmt.Scanln(&releaseYear); err != nil {
		log.Fatalf("Error scanning release year: %v", err)
	}

	for {
		fmt.Print("Actor (type 'exit' to finish): ")
		if !scanner.Scan() {
			log.Fatalf("Error reading actor: %v", scanner.Err())
		}
		actor := scanner.Text()
		if actor == "exit" {
			break
		}

		var actorID int
		err := db.QueryRow("SELECT id FROM people WHERE name = ?", actor).Scan(&actorID)
		if err != nil {
			fmt.Printf("- We could not find \"%s\", try again!\n", actor)
		} else {
			actors = append(actors, actor)
		}
	}
	totalMinutes := hours*60 + minutes
	addMovie(db, title, totalMinutes, director, releaseYear, actors)
}

func addPersonInteractive(db *sql.DB, scanner *bufio.Scanner) {
	var name string
	var birthYear int
	fmt.Print("Name: ")
	if !scanner.Scan() {
		log.Fatalf("Error reading name: %v", scanner.Err())
	}
	name = scanner.Text()

	fmt.Print("Birth Year: ")
	if _, err := fmt.Scanln(&birthYear); err != nil {
		log.Fatalf("Error scanning birth year: %v", err)
	}
	addPerson(db, name, birthYear)
}

func handleDeleteCommands(db *sql.DB, args []string, scanner *bufio.Scanner) {
	if len(args) > 0 && args[0] == "-p" {
		var name string
		fmt.Print("Name: ")
		if !scanner.Scan() {
			log.Fatalf("Error reading name: %v", scanner.Err())
		}
		name = scanner.Text()
		deletePerson(db, name)
	} else {
		fmt.Println("Invalid delete command. Use 'p' for person.")
	}
}

func printCommandDocumentation() {
	fmt.Println("============================================================================")
	fmt.Println("Available Commands:")
	fmt.Println("- List Movies: l")
	fmt.Println("  - l: List all movies alphabetically by title")
	fmt.Println("  - l -v: List movies with details including actors and ages")
	fmt.Println("  - l -t \"regex\": List movies matching the title regex")
	fmt.Println("  - l -d \"regex\": Filter movies by director matching the regex")
	fmt.Println("  - l -a \"regex\": Filter movies by actors matching the regex")
	fmt.Println("  - l -la: List movies in ascending order by length, then title")
	fmt.Println("  - l -ld: List movies in descending order by length, then title")
	fmt.Println("- Add Entries:")
	fmt.Println("  - Add Person: a -p")
	fmt.Println("    - Adds a new person to the database")
	fmt.Println("  - Add Movie: a -m")
	fmt.Println("    - Adds a new movie to the database")
	fmt.Println("- Delete Entries:")
	fmt.Println("  - Delete Person: d -p \"name\"")
	fmt.Println("    - Deletes a person from the database and their associations with movies")
	fmt.Println("- Exit Application: exit")
	fmt.Println("============================================================================")
}


// func handleAddCommands(db *sql.DB, args []string, scanner *bufio.Scanner) {
// 	if len(args) == 0 {
// 		fmt.Println("Please specify what you want to add (movie or person).")
// 		return
// 	}

// 	switch args[0] {
// 	case "-m":
// 		fmt.Println("Enter the movie details:")
// 		fmt.Print("Title: ")
// 		scanner.Scan()
// 		title := scanner.Text()
// 		fmt.Print("Length (in minutes): ")
// 		scanner.Scan()
// 		length, err := strconv.Atoi(scanner.Text())
// 		if err != nil {
// 			fmt.Println("Invalid length.")
// 			return
// 		}
// 		fmt.Print("Director: ")
// 		scanner.Scan()
// 		director := scanner.Text()
// 		fmt.Print("Release Year: ")
// 		scanner.Scan()
// 		releaseYear, err := strconv.Atoi(scanner.Text())
// 		if err != nil {
// 			fmt.Println("Invalid release year.")
// 			return
// 		}

// 		fmt.Println("Enter actors (comma separated):")
// 		scanner.Scan()
// 		actors := strings.Split(scanner.Text(), ",")

// 		addMovie(db, title, length, director, releaseYear, actors)
// 	case "-p":
// 		fmt.Println("Enter the person details:")
// 		fmt.Print("Name: ")
// 		scanner.Scan()
// 		name := scanner.Text()
// 		fmt.Print("Birth Year: ")
// 		scanner.Scan()
// 		birthYear, err := strconv.Atoi(scanner.Text())
// 		if err != nil {
// 			fmt.Println("Invalid birth year.")
// 			return
// 		}

// 		addPerson(db, name, birthYear)
// 	default:
// 		fmt.Println("Invalid add command. Use 'm' for movie or 'p' for person.")
// 	}
// }
