package database

// Utility scripts

import (
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/lib/pq"
)

const (
	deploymentFood        = 100.0
	defaultAgentSpeed     = 1.0
	defaultAgentFoodCost  = 1.0
	defaultAgentFoodOwned = 100.0
	maxDestinationCount   = 5
)

type Database struct{ db *sql.DB }

type User struct {
	UserID            int64
	Username          string
	Money             int
	MaxAgents         int
	GuildLevel        int
	GuildExperience   float64
	GuildNextLevelExp int
}

type Agent struct {
	AgentID   int64
	OwnerID   int64
	AgentName string
	ClassName string

	Speed     float64
	FoodCost  float64
	FoodOwned float64

	JourneyGold int

	CurrentHealth float64
	MaxHealth     float64
	HealthRegen   float64

	CurrentMana float64
	MaxMana     float64
	ManaRegen   float64

	Strength     float64
	Endurance    float64
	Intelligence float64
	Wisdom       float64
	Agility      float64
	Charisma     float64

	Defense float64
	Resist  float64

	CriticalChance float64
	CriticalDamage float64

	Experience   float64
	Level        int
	NextLevelExp int

	Deployed      bool
	WorldPosition *[2]float64
	HomePosition  *[2]float64
	PathIndex     int
	ReturningHome bool
}

type Destination struct {
	DestinationID    int64
	AgentID          int64
	DestinationOrder int
	DestinationName  string
	WorldPosition    [2]float64
}

type DeployedAgent struct {
	AgentID        int64
	WorldPosition  [2]float64
	HomePosition   [2]float64
	TargetPosition [2]float64
	Speed          float64
	FoodCost       float64
	FoodOwned      float64
	PathIndex      int
	ReturningHome  bool
}

type AgentMovementUpdate struct {
	AgentID       int64
	WorldPosition [2]float64
	FoodOwned     float64
	PathIndex     int
	ReturningHome bool
}

type AuthUser struct {
	UserID       int64
	Username     string
	PasswordHash string
	Money        int
}

type UserSummary struct {
	Username       string
	GuildLevel     int
	ExpToNextLevel float64
	Gold           int
	AgentCount     int
	DeployedCount  int
}

type HireState struct {
	Gold             int
	MaxAgents        int
	CurrentAgents    int
	UndeployedAgents []Agent
}

type Class struct {
	ClassName string

	BaseMaxHealth   float64
	BaseHealthRegen float64
	BaseMaxMana     float64
	BaseManaRegen   float64
	BaseFoodCost    float64
	BaseSpeed       float64

	BaseStrength     float64
	BaseEndurance    float64
	BaseIntelligence float64
	BaseWisdom       float64
	BaseAgility      float64
	BaseCharisma     float64

	BaseDefense        float64
	BaseResist         float64
	BaseCriticalChance float64
	BaseCriticalDamage float64

	PerLevelMaxHealth   float64
	PerLevelHealthRegen float64
	PerLevelMaxMana     float64
	PerLevelManaRegen   float64
	PerLevelFoodCost    float64
	PerLevelSpeed       float64

	PerLevelStrength     float64
	PerLevelEndurance    float64
	PerLevelIntelligence float64
	PerLevelWisdom       float64
	PerLevelAgility      float64
	PerLevelCharisma     float64

	PerLevelDefense        float64
	PerLevelResist         float64
	PerLevelCriticalChance float64
	PerLevelCriticalDamage float64

	BaseWeapon string
	BaseArmor  string
}

type rowScanner interface{ Scan(dest ...any) error }

func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Database{db: db}, nil
}

func (d *Database) GetUserByUsername(username string) (AuthUser, error) {
	var user AuthUser

	err := d.db.QueryRow(`
		SELECT user_id, username, passwordhash, money
		FROM users
		WHERE username = $1
	`, username).Scan(
		&user.UserID,
		&user.Username,
		&user.PasswordHash,
		&user.Money,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return AuthUser{}, sql.ErrNoRows
	}

	if err != nil {
		return AuthUser{}, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

func (d *Database) CreateUser(username, passwordHash string) (int64, error) {
	const query = `
		INSERT INTO users (
			username, passwordhash, money, max_agents,
			guild_level, guild_experience
		)
		VALUES ($1, $2, 1000, 1, 1, 0)
		RETURNING user_id
	`
	var userID int64
	if err := d.db.QueryRow(query, username, passwordHash).Scan(&userID); err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	return userID, nil
}

func (d *Database) CreateAgent(ownerID int64, name string) (int64, error) {
	const query = `
        INSERT INTO agents (
            owner_id, agentname, deployed, world_position,
			home_position, path_index, returning_home, speed, food_cost, food_owned,
			current_health, max_health, current_mana, max_mana,
			strength, endurance, intelligence, wisdom, agility, charisma,
			experience, level
        )
		VALUES (
			$1, $2, FALSE, NULL, NULL, 0, FALSE, $3, $4, $5,
			100.0, 100, 100.0, 100, 1, 1, 1, 1, 1, 1, 0.0, 1
		)
        RETURNING agent_id
    `
	var agentID int64
	if err := d.db.QueryRow(query, ownerID, name, defaultAgentSpeed, defaultAgentFoodCost, defaultAgentFoodOwned).Scan(&agentID); err != nil {
		return 0, fmt.Errorf("create agent: %w", err)
	}
	return agentID, nil
}

func (d *Database) GetUsers() ([]User, error) {
	rows, err := d.db.Query(`
		SELECT user_id, username, money, max_agents, guild_level,
			   guild_experience, guild_next_level_exp
		FROM users
		ORDER BY user_id
	`)
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.Money,
			&user.MaxAgents,
			&user.GuildLevel,
			&user.GuildExperience,
			&user.GuildNextLevelExp,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

func (d *Database) GetUserSummary(userID int64) (UserSummary, error) {
	var summary UserSummary

	err := d.db.QueryRow(`
		SELECT
			u.username,
			u.guild_level,
			GREATEST(u.guild_next_level_exp - u.guild_experience, 0),
			u.money,
			COUNT(a.agent_id),
			COUNT(a.agent_id) FILTER (WHERE a.deployed = TRUE)
		FROM users AS u
		LEFT JOIN agents AS a
			ON a.owner_id = u.user_id
		WHERE u.user_id = $1
		GROUP BY
			u.user_id,
			u.username,
			u.guild_level,
			u.guild_next_level_exp,
			u.guild_experience,
			u.money
	`, userID).Scan(
		&summary.Username,
		&summary.GuildLevel,
		&summary.ExpToNextLevel,
		&summary.Gold,
		&summary.AgentCount,
		&summary.DeployedCount,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return UserSummary{}, fmt.Errorf("user %d not found", userID)
	}

	if err != nil {
		return UserSummary{}, fmt.Errorf("get user summary: %w", err)
	}

	return summary, nil
}
func (d *Database) PrintUserSummary(userID int64) error {
	summary, err := d.GetUserSummary(userID)
	if err != nil {
		return err
	}

	fmt.Printf("\nGuild %s\n", summary.Username)
	fmt.Printf("Level %d\n", summary.GuildLevel)
	fmt.Printf("%.2f to next level\n", summary.ExpToNextLevel)
	fmt.Printf("Gold: %d\n", summary.Gold)
	fmt.Printf("Agents: %d\n", summary.AgentCount)
	fmt.Printf("Deployed: %d\n", summary.DeployedCount)

	return nil
}

func (d *Database) UpdateUserProgress(
	userID int64,
	money, maxAgents, guildLevel int,
	guildExperience float64,
) error {
	result, err := d.db.Exec(`
        UPDATE users
        SET money = $2,
            max_agents = $3,
            guild_level = $4,
            guild_experience = $5
        WHERE user_id = $1
    `, userID, money, maxAgents, guildLevel, guildExperience)
	if err != nil {
		return fmt.Errorf("update user %d: %w", userID, err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("update user %d rows affected: %w", userID, err)
	} else if affected == 0 {
		return fmt.Errorf("user %d not found", userID)
	}
	return nil
}

func (d *Database) GetAgent(agentID int64) (Agent, error) {
	agent, err := scanAgent(
		d.db.QueryRow(
			agentSelectQuery+` WHERE agent_id = $1`,
			agentID,
		),
	)
	if err != nil {
		return Agent{}, fmt.Errorf(
			"get agent %d: %w",
			agentID,
			err,
		)
	}

	return agent, nil
}
func (d *Database) GetAgents() ([]Agent, error) {
	rows, err := d.db.Query(agentSelectQuery + ` ORDER BY agent_id`)
	if err != nil {
		return nil, fmt.Errorf("get agents: %w", err)
	}
	defer rows.Close()
	return scanAgents(rows)
}

func (d *Database) UpdateAgentStats(
	agentID int64,
	currentHealth float64,
	maxHealth float64,
	currentMana float64,
	maxMana float64,
	strength float64,
	endurance float64,
	intelligence float64,
	wisdom float64,
	agility float64,
	charisma float64,
	experience float64,
	level int,
) error {
	result, err := d.db.Exec(`
        UPDATE agents
        SET current_health = $2,
            max_health = $3,
            current_mana = $4,
            max_mana = $5,
            strength = $6,
            endurance = $7,
            intelligence = $8,
            wisdom = $9,
            agility = $10,
            charisma = $11,
            experience = $12,
            level = $13
        WHERE agent_id = $1
    `, agentID, currentHealth, maxHealth, currentMana, maxMana,
		strength, endurance, intelligence, wisdom, agility, charisma,
		experience, level)
	if err != nil {
		return fmt.Errorf("update agent %d: %w", agentID, err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("update agent %d rows affected: %w", agentID, err)
	} else if affected == 0 {
		return fmt.Errorf("agent %d not found", agentID)
	}
	return nil
}

func (d *Database) GetAgentsByOwner(ownerID int64) ([]Agent, error) {
	rows, err := d.db.Query(agentSelectQuery+` WHERE owner_id = $1 ORDER BY agent_id`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("get agents for owner %d: %w", ownerID, err)
	}
	defer rows.Close()
	return scanAgents(rows)
}

func (d *Database) GetAgentByOwner(ownerID, agentID int64) (Agent, error) {
	agent, err := scanAgent(d.db.QueryRow(
		agentSelectQuery+` WHERE owner_id = $1 AND agent_id = $2`,
		ownerID,
		agentID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return Agent{}, fmt.Errorf("agent %d was not found for owner %d", agentID, ownerID)
	}
	if err != nil {
		return Agent{}, fmt.Errorf("get agent status: %w", err)
	}
	return agent, nil
}

func (d *Database) SetDestinationAndDeploy(
	ownerID, agentID int64,
	homePosition, destination [2]float64,
) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("begin set-destination transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var deployed bool
	err = tx.QueryRow(`
        SELECT deployed
        FROM agents
        WHERE agent_id = $1 AND owner_id = $2
        FOR UPDATE
    `, agentID, ownerID).Scan(&deployed)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("agent %d was not found for owner %d", agentID, ownerID)
	}
	if err != nil {
		return fmt.Errorf("load agent for deployment: %w", err)
	}
	if deployed {
		return fmt.Errorf("agent %d is already deployed", agentID)
	}

	if _, err := tx.Exec(`DELETE FROM destinations WHERE agent_id = $1`, agentID); err != nil {
		return fmt.Errorf("clear previous destinations: %w", err)
	}

	destinationName := fmt.Sprintf("(%g, %g)", destination[0], destination[1])
	if _, err := tx.Exec(`
        INSERT INTO destinations (
            agent_id, destination_order, destination_name, world_position
        )
        VALUES ($1, 0, $2, $3)
    `, agentID, destinationName, positionArray(destination)); err != nil {
		return fmt.Errorf("insert destination: %w", err)
	}

	if _, err := tx.Exec(`
        UPDATE agents
        SET
            food_owned = $2,
            deployed = TRUE,
            world_position = $3,
            home_position = $3,
            path_index = 0,
            returning_home = FALSE
        WHERE agent_id = $1
    `, agentID, deploymentFood, positionArray(homePosition)); err != nil {
		return fmt.Errorf("deploy agent: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit set-destination transaction: %w", err)
	}
	return nil
}

func (d *Database) AddDestination(agentID int64, destination [2]float64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("begin add-destination transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var deployed bool
	err = tx.QueryRow(`
        SELECT deployed
        FROM agents
        WHERE agent_id = $1
        FOR UPDATE
    `, agentID).Scan(&deployed)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("agent %d was not found", agentID)
	}
	if err != nil {
		return fmt.Errorf("load agent for destination append: %w", err)
	}
	if !deployed {
		return fmt.Errorf("agent %d is not deployed", agentID)
	}

	var count int
	err = tx.QueryRow(`
        SELECT COUNT(*)
        FROM destinations
        WHERE agent_id = $1
    `, agentID).Scan(&count)
	if err != nil {
		return fmt.Errorf("count existing destinations: %w", err)
	}
	if count >= maxDestinationCount {
		return fmt.Errorf("agent %d already has the maximum of %d destinations", agentID, maxDestinationCount)
	}

	destinationName := fmt.Sprintf("(%g, %g)", destination[0], destination[1])
	if _, err := tx.Exec(`
        INSERT INTO destinations (
            agent_id,
            destination_order,
            destination_name,
            world_position
        )
        VALUES ($1, $2, $3, $4)
    `, agentID, count, destinationName, positionArray(destination)); err != nil {
		return fmt.Errorf("append destination: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit add-destination transaction: %w", err)
	}
	return nil
}

func (d *Database) WithdrawAgent(agentID int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("begin withdraw-agent transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Exec(`
        UPDATE agents
        SET
            deployed = FALSE,
            food_owned = 0.0,
            world_position = NULL,
            home_position = NULL,
            path_index = 0,
            returning_home = FALSE
        WHERE agent_id = $1
    `, agentID)
	if err != nil {
		return fmt.Errorf("withdraw agent: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("withdraw agent rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("agent %d not found", agentID)
	}

	if _, err := tx.Exec(`DELETE FROM destinations WHERE agent_id = $1`, agentID); err != nil {
		return fmt.Errorf("clear agent destinations: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit withdraw-agent transaction: %w", err)
	}
	return nil
}

func (d *Database) GetDestinationsByAgent(agentID int64) ([]Destination, error) {
	rows, err := d.db.Query(`
        SELECT
            destination_id,
            agent_id,
            destination_order,
            destination_name,
            world_position[1],
            world_position[2]
        FROM destinations
        WHERE agent_id = $1
        ORDER BY destination_order
    `, agentID)
	if err != nil {
		return nil, fmt.Errorf("get destinations for agent %d: %w", agentID, err)
	}
	defer rows.Close()

	destinations := make([]Destination, 0)
	for rows.Next() {
		var destination Destination
		if err := rows.Scan(
			&destination.DestinationID,
			&destination.AgentID,
			&destination.DestinationOrder,
			&destination.DestinationName,
			&destination.WorldPosition[0],
			&destination.WorldPosition[1],
		); err != nil {
			return nil, fmt.Errorf("scan destination: %w", err)
		}
		destinations = append(destinations, destination)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate destinations: %w", err)
	}
	return destinations, nil
}

func (d *Database) GetDeployedAgents() ([]DeployedAgent, error) {
	rows, err := d.db.Query(`
        SELECT
            a.agent_id,
            a.world_position[1],
            a.world_position[2],
            a.home_position[1],
            a.home_position[2],
            CASE WHEN a.returning_home THEN a.home_position[1] ELSE d.world_position[1] END,
            CASE WHEN a.returning_home THEN a.home_position[2] ELSE d.world_position[2] END,
            a.speed,
            a.food_cost,
            a.food_owned,
            a.path_index,
            a.returning_home
        FROM agents AS a
        LEFT JOIN destinations AS d
            ON d.agent_id = a.agent_id
           AND d.destination_order = a.path_index
        WHERE a.deployed = TRUE
          AND a.world_position IS NOT NULL
          AND a.home_position IS NOT NULL
          AND (a.returning_home = TRUE OR d.destination_id IS NOT NULL)
        ORDER BY a.agent_id
    `)
	if err != nil {
		return nil, fmt.Errorf("get deployed agents: %w", err)
	}
	defer rows.Close()

	agents := make([]DeployedAgent, 0)
	for rows.Next() {
		var agent DeployedAgent
		if err := rows.Scan(
			&agent.AgentID,
			&agent.WorldPosition[0],
			&agent.WorldPosition[1],
			&agent.HomePosition[0],
			&agent.HomePosition[1],
			&agent.TargetPosition[0],
			&agent.TargetPosition[1],
			&agent.Speed,
			&agent.FoodCost,
			&agent.FoodOwned,
			&agent.PathIndex,
			&agent.ReturningHome,
		); err != nil {
			return nil, fmt.Errorf("scan deployed agent: %w", err)
		}
		agents = append(agents, agent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate deployed agents: %w", err)
	}
	return agents, nil
}
func (d *Database) ApplyAgentMovement(update AgentMovementUpdate) error {
	result, err := d.db.Exec(`
		UPDATE agents
		SET
			experience = experience + (
				SQRT(
					POWER(($2::DOUBLE PRECISION[])[1] - world_position[1], 2) +
					POWER(($2::DOUBLE PRECISION[])[2] - world_position[2], 2)
				) * 0.1
			),
			world_position = $2,
			food_owned = $3,
			path_index = $4,
			returning_home = $5
		WHERE agent_id = $1
		  AND deployed = TRUE
		  AND world_position IS NOT NULL
	`,
		update.AgentID,
		positionArray(update.WorldPosition),
		update.FoodOwned,
		update.PathIndex,
		update.ReturningHome,
	)
	if err != nil {
		return fmt.Errorf(
			"apply movement for agent %d: %w",
			update.AgentID,
			err,
		)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"movement rows affected for agent %d: %w",
			update.AgentID,
			err,
		)
	}

	if affected == 0 {
		return fmt.Errorf(
			"agent %d is no longer deployed",
			update.AgentID,
		)
	}

	return nil
}

func (d *Database) UpdateAgentPosition(
	agentID int64,
	newPosition [2]float64,
) error {
	result, err := d.db.Exec(`
		UPDATE agents
		SET
			experience = experience + (
				SQRT(
					POWER(($2::DOUBLE PRECISION[])[1] - world_position[1], 2) +
					POWER(($2::DOUBLE PRECISION[])[2] - world_position[2], 2)
				) * 0.01
			),
			world_position = $2
		WHERE agent_id = $1
		  AND deployed = TRUE
		  AND world_position IS NOT NULL
	`,
		agentID,
		positionArray(newPosition),
	)
	if err != nil {
		return fmt.Errorf(
			"update agent position: %w",
			err,
		)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"update agent position rows affected: %w",
			err,
		)
	}

	if affected == 0 {
		return fmt.Errorf(
			"agent %d not found or is not deployed",
			agentID,
		)
	}

	return nil
}

func (d *Database) BeginReturnHome(agentID int64) error {
	result, err := d.db.Exec(`
        UPDATE agents
        SET returning_home = TRUE
        WHERE agent_id = $1
          AND deployed = TRUE
          AND returning_home = FALSE
    `, agentID)
	if err != nil {
		return fmt.Errorf("begin return home: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("begin return home rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("agent %d not found, not deployed, or already returning", agentID)
	}
	return nil
}

func (d *Database) CompleteJourney(agentID int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("begin complete-journey transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Exec(`
		UPDATE agents
		SET
			deployed = FALSE,
			food_owned = 0.0,

			current_health = max_health,
			current_mana = max_mana,

			world_position = NULL,
			home_position = NULL,
			path_index = 0,
			returning_home = FALSE
		WHERE agent_id = $1
    `, agentID)
	if err != nil {
		return fmt.Errorf("complete journey: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("complete journey rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("agent %d is not completing a return journey", agentID)
	}

	if _, err := tx.Exec(`DELETE FROM destinations WHERE agent_id = $1`, agentID); err != nil {
		return fmt.Errorf("clear completed journey destinations: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit complete-journey transaction: %w", err)
	}
	return nil
}

func (d *Database) LogEvent(agentID int64, title, description, resultText string) error {
	var eventID int64
	err := d.db.QueryRow(`
        INSERT INTO event_results (
            agent_id, title, description, results, where_position
        )
        SELECT agent_id, $2, $3, $4, world_position
        FROM agents
        WHERE agent_id = $1
          AND deployed = TRUE
          AND world_position IS NOT NULL
        RETURNING event_id
    `, agentID, title, description, resultText).Scan(&eventID)

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("agent %d not found or is not deployed", agentID)
	}
	if err != nil {
		return fmt.Errorf("log event: %w", err)
	}
	return nil
}

func (d *Database) GetHireState(userID int64) (HireState, error) {
	var state HireState

	err := d.db.QueryRow(`
		SELECT
			u.money,
			u.max_agents,
			COUNT(a.agent_id)
		FROM users AS u
		LEFT JOIN agents AS a
			ON a.owner_id = u.user_id
		WHERE u.user_id = $1
		GROUP BY u.user_id, u.money, u.max_agents
	`, userID).Scan(
		&state.Gold,
		&state.MaxAgents,
		&state.CurrentAgents,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return HireState{}, fmt.Errorf("user %d not found", userID)
	}
	if err != nil {
		return HireState{}, fmt.Errorf("get hire state: %w", err)
	}

	agents, err := d.GetAgentsByOwner(userID)
	if err != nil {
		return HireState{}, err
	}

	for _, agent := range agents {
		if !agent.Deployed {
			state.UndeployedAgents = append(state.UndeployedAgents, agent)
		}
	}

	return state, nil
}

func (d *Database) ReleaseHero(userID, agentID int64) error {
	result, err := d.db.Exec(`
		DELETE FROM agents
		WHERE agent_id = $1
		  AND owner_id = $2
		  AND deployed = FALSE
	`, agentID, userID)
	if err != nil {
		return fmt.Errorf("release hero: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("release hero rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("hero does not exist, does not belong to this user, or is deployed")
	}

	return nil
}

func (d *Database) HireHero(
	userID int64,
	heroName string,
	className string,
) (int64, error) {
	// const hireCost = 600 // Depreciated

	switch className {
	case "Warrior", "Rogue", "Ranger", "Wizard", "Cleric":
	default:
		return 0, fmt.Errorf("class %q cannot be hired", className)
	}

	class, err := d.GetClass(className)
	if err != nil {
		return 0, err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin hire transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var money, maxAgents int
	err = tx.QueryRow(`
		SELECT money, max_agents
		FROM users
		WHERE user_id = $1
		FOR UPDATE
	`, userID).Scan(&money, &maxAgents)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("user %d not found", userID)
	}
	if err != nil {
		return 0, fmt.Errorf("load user for hire: %w", err)
	}

	var currentAgents int
	if err := tx.QueryRow(`
		SELECT COUNT(*)
		FROM agents
		WHERE owner_id = $1
	`, userID).Scan(&currentAgents); err != nil {
		return 0, fmt.Errorf("count heroes: %w", err)
	}
	if currentAgents >= 5 {
		return 0, fmt.Errorf("hero capacity reached")
	}

	hireCost := 0

	if currentAgents > 0 {
		hireCost = 500 + (currentAgents * 100)
	}

	if currentAgents >= maxAgents {
		return 0, fmt.Errorf("hero capacity reached")
	}
	if currentAgents > 0 && money < hireCost {
		return 0, fmt.Errorf("not enough Gold")
	}

	if heroName == "" {
		heroName = className
	}

	var agentID int64
	err = tx.QueryRow(`
		INSERT INTO agents (
			owner_id,
			agentname,
			class_name,

			deployed,
			world_position,
			home_position,
			path_index,
			returning_home,

			speed,
			food_cost,
			food_owned,
			journey_gold,

			current_health,
			max_health,
			health_regen,

			current_mana,
			max_mana,
			mana_regen,

			strength,
			endurance,
			intelligence,
			wisdom,
			agility,
			charisma,

			defense,
			resist,

			critical_chance,
			critical_damage,

			experience,
			level
		)
		VALUES (
			$1, $2, $3,

			FALSE, NULL, NULL, 0, FALSE,

			$4, $5, 0.0, 0,

			$6, $6, $7,

			$8, $8, $9,

			$10, $11, $12, $13, $14, $15,

			$16, $17,

			$18, $19,

			0.0, 1
		)
		RETURNING agent_id
	`,
		userID,
		heroName,
		class.ClassName,

		class.BaseSpeed,
		class.BaseFoodCost,

		class.BaseMaxHealth,
		class.BaseHealthRegen,

		class.BaseMaxMana,
		class.BaseManaRegen,

		class.BaseStrength,
		class.BaseEndurance,
		class.BaseIntelligence,
		class.BaseWisdom,
		class.BaseAgility,
		class.BaseCharisma,

		class.BaseDefense,
		class.BaseResist,

		class.BaseCriticalChance,
		class.BaseCriticalDamage,
	).Scan(&agentID)
	if err != nil {
		return 0, fmt.Errorf("create hero: %w", err)
	}

	if class.BaseWeapon != "" {
		if _, err := tx.Exec(`
			INSERT INTO agent_equipment (
				agent_id,
				equipment_name,
				equipment_slot
			)
			VALUES ($1, $2, 'Weapon')
		`, agentID, class.BaseWeapon); err != nil {
			return 0, fmt.Errorf("equip starting weapon: %w", err)
		}
	}

	if class.BaseArmor != "" {
		if _, err := tx.Exec(`
			INSERT INTO agent_equipment (
				agent_id,
				equipment_name,
				equipment_slot
			)
			VALUES ($1, $2, 'Armor')
		`, agentID, class.BaseArmor); err != nil {
			return 0, fmt.Errorf("equip starting armor: %w", err)
		}
	}

	// Commits the cost and completes the purchase
	if _, err := tx.Exec(`
		UPDATE users
		SET money = money - $2
		WHERE user_id = $1
	`, userID, hireCost); err != nil {
		return 0, fmt.Errorf("deduct hire cost: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit hire transaction: %w", err)
	}

	return agentID, nil
}

func routeDistance(
	home [2]float64,
	destinations [][2]float64,
) float64 {
	if len(destinations) == 0 {
		return 0
	}

	total := 0.0
	current := home

	for _, destination := range destinations {
		total += math.Hypot(
			destination[0]-current[0],
			destination[1]-current[1],
		)
		current = destination
	}

	// After the final destination, the Hero returns directly home.
	total += math.Hypot(
		home[0]-current[0],
		home[1]-current[1],
	)

	return total
}

func (d *Database) DeployAgentRoute(
	userID int64,
	agentID int64,
	homePosition [2]float64,
	destinations [][2]float64,
	additionalFood int,
) error {
	if len(destinations) == 0 {
		return fmt.Errorf("at least one destination is required")
	}
	if len(destinations) > maxDestinationCount {
		return fmt.Errorf(
			"a maximum of %d destinations is allowed",
			maxDestinationCount,
		)
	}
	if additionalFood < 0 {
		return fmt.Errorf("additional food cannot be negative")
	}

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("begin deployment transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var deployed bool
	var gold int

	err = tx.QueryRow(`
		SELECT
			a.deployed,
			u.money
		FROM agents AS a
		JOIN users AS u
			ON u.user_id = a.owner_id
		WHERE a.agent_id = $1
		AND a.owner_id = $2
		FOR UPDATE
	`, agentID, userID).Scan(
		&deployed,
		&gold,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("hero was not found")
	}
	if err != nil {
		return fmt.Errorf("load hero for deployment: %w", err)
	}

	if deployed {
		return fmt.Errorf("hero is already deployed")
	}
	if additionalFood > gold {
		return fmt.Errorf("not enough Gold for additional food")
	}
	totalFood := deploymentFood + float64(additionalFood)

	if _, err := tx.Exec(
		`DELETE FROM destinations WHERE agent_id = $1`,
		agentID,
	); err != nil {
		return fmt.Errorf("clear previous destinations: %w", err)
	}

	for index, destination := range destinations {
		destinationName := fmt.Sprintf(
			"(%.2f, %.2f)",
			destination[0],
			destination[1],
		)

		if _, err := tx.Exec(`
			INSERT INTO destinations (
				agent_id,
				destination_order,
				destination_name,
				world_position
			)
			VALUES ($1, $2, $3, $4)
		`,
			agentID,
			index,
			destinationName,
			positionArray(destination),
		); err != nil {
			return fmt.Errorf(
				"insert destination %d: %w",
				index+1,
				err,
			)
		}
	}

	if _, err := tx.Exec(`
		UPDATE agents
		SET
			deployed = TRUE,
			world_position = $2,
			home_position = $2,
			path_index = 0,
			returning_home = FALSE,
			food_owned = $3,
			journey_gold = 0
		WHERE agent_id = $1
	`,
		agentID,
		positionArray(homePosition),
		totalFood,
	); err != nil {
		return fmt.Errorf("deploy hero: %w", err)
	}

	if additionalFood > 0 {
		if _, err := tx.Exec(`
			UPDATE users
			SET money = money - $2
			WHERE user_id = $1
		`, userID, additionalFood); err != nil {
			return fmt.Errorf("purchase additional food: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit deployment transaction: %w", err)
	}

	return nil
}

func (d *Database) ExecSchema(schema string) (sql.Result, error) {
	result, err := d.db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("execute schema: %w", err)
	}
	return result, nil
}

func (d *Database) Close() error { return d.db.Close() }

const agentSelectQuery = `
	SELECT
		agent_id,
		owner_id,
		agentname,
		COALESCE(class_name, ''),

		speed,
		food_cost,
		food_owned,
		journey_gold,

		current_health,
		max_health,
		health_regen,

		current_mana,
		max_mana,
		mana_regen,

		strength,
		endurance,
		intelligence,
		wisdom,
		agility,
		charisma,

		defense,
		resist,

		critical_chance,
		critical_damage,

		experience,
		level,
		next_level_exp,

		deployed,
		world_position[1],
		world_position[2],
		home_position[1],
		home_position[2],
		path_index,
		returning_home
	FROM agents
`

func scanAgents(rows *sql.Rows) ([]Agent, error) {
	agents := make([]Agent, 0)
	for rows.Next() {
		agent, err := scanAgent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		agents = append(agents, agent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate agents: %w", err)
	}
	return agents, nil
}

func scanAgent(scanner rowScanner) (Agent, error) {
	var agent Agent
	var worldX, worldY, homeX, homeY sql.NullFloat64

	if err := scanner.Scan(
		&agent.AgentID,
		&agent.OwnerID,
		&agent.AgentName,
		&agent.ClassName,

		&agent.Speed,
		&agent.FoodCost,
		&agent.FoodOwned,
		&agent.JourneyGold,

		&agent.CurrentHealth,
		&agent.MaxHealth,
		&agent.HealthRegen,

		&agent.CurrentMana,
		&agent.MaxMana,
		&agent.ManaRegen,

		&agent.Strength,
		&agent.Endurance,
		&agent.Intelligence,
		&agent.Wisdom,
		&agent.Agility,
		&agent.Charisma,

		&agent.Defense,
		&agent.Resist,

		&agent.CriticalChance,
		&agent.CriticalDamage,

		&agent.Experience,
		&agent.Level,
		&agent.NextLevelExp,

		&agent.Deployed,
		&worldX,
		&worldY,
		&homeX,
		&homeY,
		&agent.PathIndex,
		&agent.ReturningHome,
	); err != nil {
		return Agent{}, err
	}

	if worldX.Valid && worldY.Valid {
		position := [2]float64{worldX.Float64, worldY.Float64}
		agent.WorldPosition = &position
	}

	if homeX.Valid && homeY.Valid {
		position := [2]float64{homeX.Float64, homeY.Float64}
		agent.HomePosition = &position
	}

	return agent, nil
}

func positionArray(position [2]float64) interface{} {
	return pq.Array([]float64{position[0], position[1]})
}
func (d *Database) GetClass(className string) (Class, error) {
	var class Class

	err := d.db.QueryRow(`
		SELECT
			class_name,

			base_max_health,
			base_health_regen,
			base_max_mana,
			base_mana_regen,
			base_food_cost,
			base_speed,

			base_strength,
			base_endurance,
			base_intelligence,
			base_wisdom,
			base_agility,
			base_charisma,

			base_defense,
			base_resist,
			base_critical_chance,
			base_critical_damage,

			per_level_max_health,
			per_level_health_regen,
			per_level_max_mana,
			per_level_mana_regen,
			per_level_food_cost,
			per_level_speed,

			per_level_strength,
			per_level_endurance,
			per_level_intelligence,
			per_level_wisdom,
			per_level_agility,
			per_level_charisma,

			per_level_defense,
			per_level_resist,
			per_level_critical_chance,
			per_level_critical_damage,

			COALESCE(base_weapon, ''),
			COALESCE(base_armor_equipment, '')

		FROM classes
		WHERE class_name = $1
	`, className).Scan(
		&class.ClassName,

		&class.BaseMaxHealth,
		&class.BaseHealthRegen,
		&class.BaseMaxMana,
		&class.BaseManaRegen,
		&class.BaseFoodCost,
		&class.BaseSpeed,

		&class.BaseStrength,
		&class.BaseEndurance,
		&class.BaseIntelligence,
		&class.BaseWisdom,
		&class.BaseAgility,
		&class.BaseCharisma,

		&class.BaseDefense,
		&class.BaseResist,
		&class.BaseCriticalChance,
		&class.BaseCriticalDamage,

		&class.PerLevelMaxHealth,
		&class.PerLevelHealthRegen,
		&class.PerLevelMaxMana,
		&class.PerLevelManaRegen,
		&class.PerLevelFoodCost,
		&class.PerLevelSpeed,

		&class.PerLevelStrength,
		&class.PerLevelEndurance,
		&class.PerLevelIntelligence,
		&class.PerLevelWisdom,
		&class.PerLevelAgility,
		&class.PerLevelCharisma,

		&class.PerLevelDefense,
		&class.PerLevelResist,
		&class.PerLevelCriticalChance,
		&class.PerLevelCriticalDamage,

		&class.BaseWeapon,
		&class.BaseArmor,
	)

	if err != nil {
		return Class{}, fmt.Errorf(
			"get class %q: %w",
			className,
			err,
		)
	}

	return class, nil
}

func (d *Database) SaveAgentLevelUp(agent *Agent) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}

	result, err := d.db.Exec(`
		UPDATE agents
		SET
			max_health = $2,
			health_regen = $3,

			max_mana = $4,
			mana_regen = $5,

			strength = $6,
			endurance = $7,
			intelligence = $8,
			wisdom = $9,
			agility = $10,
			charisma = $11,

			defense = $12,
			resist = $13,

			critical_chance = $14,
			critical_damage = $15,

			speed = $16,
			food_cost = $17,

			level = $18
		WHERE agent_id = $1
	`,
		agent.AgentID,

		agent.MaxHealth,
		agent.HealthRegen,

		agent.MaxMana,
		agent.ManaRegen,

		agent.Strength,
		agent.Endurance,
		agent.Intelligence,
		agent.Wisdom,
		agent.Agility,
		agent.Charisma,

		agent.Defense,
		agent.Resist,

		agent.CriticalChance,
		agent.CriticalDamage,

		agent.Speed,
		agent.FoodCost,

		agent.Level,
	)
	if err != nil {
		return fmt.Errorf(
			"save level-up for agent %d: %w",
			agent.AgentID,
			err,
		)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"save level-up rows affected for agent %d: %w",
			agent.AgentID,
			err,
		)
	}

	if affected == 0 {
		return fmt.Errorf(
			"agent %d not found",
			agent.AgentID,
		)
	}

	return nil
}
func (d *Database) SaveAgent(agent *Agent) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}

	result, err := d.db.Exec(`
		UPDATE agents
		SET
			current_health = $2,
			current_mana = $3,
			food_owned = $4,
			journey_gold = $5,
			experience = $6
		WHERE agent_id = $1
	`,
		agent.AgentID,
		agent.CurrentHealth,
		agent.CurrentMana,
		agent.FoodOwned,
		agent.JourneyGold,
		agent.Experience,
	)
	if err != nil {
		return fmt.Errorf(
			"save agent %d: %w",
			agent.AgentID,
			err,
		)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"save agent rows affected for agent %d: %w",
			agent.AgentID,
			err,
		)
	}

	if affected == 0 {
		return fmt.Errorf(
			"agent %d not found",
			agent.AgentID,
		)
	}

	return nil
}
