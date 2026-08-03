package database

import (
	"database/sql"
	"fmt"

	pq "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) CreateUser(username, passwordHash string) (int64, error) {
	var userID int64
	query := `INSERT INTO users (username, passwordhash) VALUES ($1, $2) RETURNING user_id`
	if err := d.db.QueryRow(query, username, passwordHash).Scan(&userID); err != nil {
		return 0, err
	}
	return userID, nil
}

func (d *Database) CreateAgent(ownerID int64, name string) (int64, error) {
	var agentID int64
	query := `INSERT INTO agents (owner_id, agentname, world_position) VALUES ($1, $2, $3) RETURNING agent_id`
	if err := d.db.QueryRow(query, ownerID, name, pq.Array([]float64{0, 0})).Scan(&agentID); err != nil {
		return 0, err
	}
	return agentID, nil
}

func (d *Database) AddDestination(agentID int64, position [2]float64) (int64, error) {
	var destinationID int64
	positionLabel := fmt.Sprintf("(%g, %g)", position[0], position[1])
	query := `INSERT INTO destinations (agent_id, destination_order, destination_name) VALUES ($1, 0, $2) RETURNING destination_id`
	if err := d.db.QueryRow(query, agentID, positionLabel).Scan(&destinationID); err != nil {
		return 0, err
	}
	return destinationID, nil
}

func (d *Database) UpdateAgentPosition(agentID int64, newPosition [2]float64) error {
	query := `UPDATE agents SET world_position = $2 WHERE agent_id = $1`
	result, err := d.db.Exec(query, agentID, pq.Array([]float64{newPosition[0], newPosition[1]}))
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("agent %d not found", agentID)
	}
	return nil
}

func (d *Database) LogEvent(agentID int64, title, description, result string) error {
	var currentPosition [2]float64
	if err := d.db.QueryRow(`SELECT world_position[1], world_position[2] FROM agents WHERE agent_id = $1`, agentID).Scan(&currentPosition[0], &currentPosition[1]); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("agent %d not found", agentID)
		}
		return err
	}

	_, err := d.db.Exec(`INSERT INTO event_results (agent_id, title, description, results, where_position) VALUES ($1, $2, $3, $4, $5)`, agentID, title, description, result, pq.Array([]float64{currentPosition[0], currentPosition[1]}))
	return err
}

func (d *Database) ExecSchema(schema string) (sql.Result, error) {
	return d.db.Exec(schema)
}

func (d *Database) Close() error {
	return d.db.Close()
}
