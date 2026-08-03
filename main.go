package littleadventuremap

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"littleadventuremap/database"

	_ "github.com/lib/pq"
)

func loadDatabaseURL() string {
	if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
		return connStr
	}

	file, err := os.Open(".env")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key != "DATABASE_URL" {
			continue
		}

		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		return value
	}

	if err := scanner.Err(); err != nil {
		log.Printf("warning: failed to read .env: %v", err)
	}

	return ""
}

func main() {
	connStr := loadDatabaseURL()
	if connStr == "" {
		log.Fatal("DATABASE_URL is not set; provide it in the environment or in an uncommitted .env file")
	}

	db, err := database.NewDatabase(connStr)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("failed to read schema.sql: %v", err)
	}

	if _, err := db.ExecSchema(string(schema)); err != nil {
		log.Fatalf("failed to execute schema: %v", err)
	}

	fmt.Println("Database schema created successfully")
}
