package simulation

// Utility scripts for running the database simulation

import (
	"fmt"
	"math"
	"math/rand"

	"littleadventuremap/database"
)

type AgentState struct {
	AgentID        int64
	Position       [2]float64
	HomePosition   [2]float64
	Destinations   [][2]float64
	PathIndex      int
	ReturningHome  bool
	Speed          float64
	FoodOwned      float64
	FoodCost       float64
	Handled        bool
	NearbyAgentIDs []int64
}

type TickState struct {
	Agents map[int64]*AgentState
}

type MovementResult struct {
	AgentID            int64
	Position           [2]float64
	FoodOwned          float64
	PathIndex          int
	ReturningHome      bool
	JourneyDone        bool
	Withdrawn          bool
	ReachedDestination bool
}

type AgentSnapshot struct {
	AgentID                 int64
	OwnerID                 int64
	AgentName               string
	ClassName               string
	Position                *[2]float64
	HomePosition            *[2]float64
	Destinations            [][2]float64
	PathIndex               int
	ReturningHome           bool
	Speed                   float64
	FoodCost                float64
	FoodOwned               float64
	JourneyGold             int
	CurrentHealth           float64
	MaxHealth               float64
	HealthRegen             float64
	CurrentMana             float64
	MaxMana                 float64
	ManaRegen               float64
	Level                   int
	Experience              float64
	NextLevelExp            int
	Strength                float64
	Endurance               float64
	Intelligence            float64
	Wisdom                  float64
	Agility                 float64
	Charisma                float64
	Defense                 float64
	Resist                  float64
	CriticalChance          float64
	CriticalDamage          float64
	Deployed                bool
	DestinationCount        int
	CurrentDestinationIndex int
	CurrentDestination      *[2]float64
	Heading                 string
}

type SimulationSnapshot struct {
	TickNumber int
	Agents     []AgentSnapshot
}

var (
	currentTickNumber int
	currentSnapshot   SimulationSnapshot
)

func CurrentTick() int {
	return currentTickNumber
}

func CurrentSnapshot() SimulationSnapshot {
	return currentSnapshot
}

func BuildSnapshot(db *database.Database) (SimulationSnapshot, error) {
	agents, err := db.GetAgents()
	if err != nil {
		return SimulationSnapshot{}, fmt.Errorf("load agent snapshot: %w", err)
	}

	snapshot := SimulationSnapshot{
		Agents: make([]AgentSnapshot, 0, len(agents)),
	}

	for _, agent := range agents {
		destinations, err := db.GetDestinationsByAgent(agent.AgentID)
		if err != nil {
			return SimulationSnapshot{}, fmt.Errorf("load destinations for agent %d: %w", agent.AgentID, err)
		}

		destinationCount := len(destinations)
		destinationPositions := make([][2]float64, 0, destinationCount)
		for _, destination := range destinations {
			destinationPositions = append(destinationPositions, destination.WorldPosition)
		}

		currentDestinationIndex := -1
		var currentDestination *[2]float64
		heading := "Idle"

		if agent.Deployed {
			if agent.ReturningHome {
				heading = "Returning Home"
			} else {
				heading = "Outbound"
			}

			if destinationCount > 0 {
				if agent.ReturningHome && agent.HomePosition != nil {
					position := [2]float64{(*agent.HomePosition)[0], (*agent.HomePosition)[1]}
					currentDestination = &position
					currentDestinationIndex = -1
				} else {
					currentDestinationIndex = agent.PathIndex
					if currentDestinationIndex >= 0 && currentDestinationIndex < destinationCount {
						position := destinations[currentDestinationIndex].WorldPosition
						currentDestination = &position
					}
				}
			}
		}

		snapshot.Agents = append(snapshot.Agents, AgentSnapshot{
			AgentID:                 agent.AgentID,
			OwnerID:                 agent.OwnerID,
			AgentName:               agent.AgentName,
			ClassName:               agent.ClassName,
			Position:                agent.WorldPosition,
			HomePosition:            agent.HomePosition,
			Destinations:            destinationPositions,
			PathIndex:               agent.PathIndex,
			ReturningHome:           agent.ReturningHome,
			Speed:                   agent.Speed,
			FoodCost:                agent.FoodCost,
			FoodOwned:               agent.FoodOwned,
			JourneyGold:             agent.JourneyGold,
			CurrentHealth:           agent.CurrentHealth,
			MaxHealth:               agent.MaxHealth,
			HealthRegen:             agent.HealthRegen,
			CurrentMana:             agent.CurrentMana,
			MaxMana:                 agent.MaxMana,
			ManaRegen:               agent.ManaRegen,
			Level:                   agent.Level,
			Experience:              agent.Experience,
			NextLevelExp:            agent.NextLevelExp,
			Strength:                agent.Strength,
			Endurance:               agent.Endurance,
			Intelligence:            agent.Intelligence,
			Wisdom:                  agent.Wisdom,
			Agility:                 agent.Agility,
			Charisma:                agent.Charisma,
			Defense:                 agent.Defense,
			Resist:                  agent.Resist,
			CriticalChance:          agent.CriticalChance,
			CriticalDamage:          agent.CriticalDamage,
			Deployed:                agent.Deployed,
			DestinationCount:        destinationCount,
			CurrentDestinationIndex: currentDestinationIndex,
			CurrentDestination:      currentDestination,
			Heading:                 heading,
		})
	}

	return snapshot, nil
}

func refreshSnapshot(db *database.Database) error {
	snapshot, err := BuildSnapshot(db)
	if err != nil {
		return err
	}

	currentTickNumber++
	snapshot.TickNumber = currentTickNumber
	currentSnapshot = snapshot

	return nil
}

func loadTickState(
	db *database.Database,
) (*TickState, error) {
	deployedAgents, err := db.GetDeployedAgents()
	if err != nil {
		return nil, fmt.Errorf(
			"load deployed agents: %w",
			err,
		)
	}

	state := &TickState{
		Agents: make(
			map[int64]*AgentState,
			len(deployedAgents),
		),
	}

	for _, deployedAgent := range deployedAgents {
		destinationRows, err := db.GetDestinationsByAgent(
			deployedAgent.AgentID,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"load destinations for agent %d: %w",
				deployedAgent.AgentID,
				err,
			)
		}

		destinations := make(
			[][2]float64,
			0,
			len(destinationRows),
		)

		for _, destination := range destinationRows {
			destinations = append(
				destinations,
				destination.WorldPosition,
			)
		}

		state.Agents[deployedAgent.AgentID] = &AgentState{
			AgentID:        deployedAgent.AgentID,
			Position:       deployedAgent.WorldPosition,
			HomePosition:   deployedAgent.HomePosition,
			Destinations:   destinations,
			PathIndex:      deployedAgent.PathIndex,
			ReturningHome:  deployedAgent.ReturningHome,
			Speed:          deployedAgent.Speed,
			FoodOwned:      deployedAgent.FoodOwned,
			FoodCost:       deployedAgent.FoodCost,
			Handled:        false,
			NearbyAgentIDs: make([]int64, 0),
		}
	}

	return state, nil
}

func distanceToTarget(
	agent *AgentState,
	targetPoint [2]float64,
) float64 {
	return math.Hypot(
		targetPoint[0]-agent.Position[0],
		targetPoint[1]-agent.Position[1],
	)
}

func currentTarget(
	agent *AgentState,
) ([2]float64, error) {
	if agent.ReturningHome {
		return agent.HomePosition, nil
	}

	if agent.PathIndex < 0 || agent.PathIndex >= len(agent.Destinations) {
		return [2]float64{}, fmt.Errorf(
			"outbound agent %d has invalid path index %d",
			agent.AgentID,
			agent.PathIndex,
		)
	}

	return agent.Destinations[agent.PathIndex], nil
}

func moveTowardTarget(
	agent *AgentState,
	targetPoint [2]float64,
) bool {
	distance := distanceToTarget(
		agent,
		targetPoint,
	)

	if distance <= agent.Speed {
		agent.Position = targetPoint
		return true
	}

	directionX :=
		(targetPoint[0] - agent.Position[0]) /
			distance

	directionY :=
		(targetPoint[1] - agent.Position[1]) /
			distance

	agent.Position[0] += directionX * agent.Speed
	agent.Position[1] += directionY * agent.Speed

	return false
}

func advanceRoute(
	agent *AgentState,
) {
	agent.PathIndex++

	if agent.PathIndex >= len(agent.Destinations) {
		agent.PathIndex = len(agent.Destinations)
		agent.ReturningHome = true
	}
}

func moveAgent(
	agent *AgentState,
) (MovementResult, error) {
	targetPoint, err := currentTarget(agent)
	if err != nil {
		return MovementResult{}, err
	}

	reachedTarget := moveTowardTarget(
		agent,
		targetPoint,
	)

	agent.FoodOwned -= agent.FoodCost

	journeyDone := false
	withdrawn := false

	// Only count outbound targets as destinations.
	reachedDestination := reachedTarget && !agent.ReturningHome

	if reachedTarget {
		if agent.ReturningHome {
			journeyDone = true
		} else {
			advanceRoute(agent)
		}
	}

	return MovementResult{
		AgentID:            agent.AgentID,
		Position:           agent.Position,
		FoodOwned:          agent.FoodOwned,
		PathIndex:          agent.PathIndex,
		ReturningHome:      agent.ReturningHome,
		JourneyDone:        journeyDone,
		Withdrawn:          withdrawn,
		ReachedDestination: reachedDestination,
	}, nil
}

func runMovementPhase(
	state *TickState,
) ([]MovementResult, error) {
	results := make(
		[]MovementResult,
		0,
		len(state.Agents),
	)

	for _, agent := range state.Agents {
		agent.Handled = false
		agent.NearbyAgentIDs =
			agent.NearbyAgentIDs[:0]

		result, err := moveAgent(agent)
		if err != nil {
			return nil, fmt.Errorf(
				"move agent %d: %w",
				agent.AgentID,
				err,
			)
		}

		results = append(results, result)
	}

	return results, nil
}

func saveMovementResults(
	db *database.Database,
	results []MovementResult,
) error {
	for _, result := range results {
		if result.JourneyDone {
			if result.Withdrawn {
				if err := db.WithdrawAgent(result.AgentID); err != nil {
					return fmt.Errorf(
						"withdraw agent %d: %w",
						result.AgentID,
						err,
					)
				}
			} else {
				agent, err := db.GetAgent(result.AgentID)
				if err != nil {
					return fmt.Errorf(
						"load agent %d for level up: %w",
						result.AgentID,
						err,
					)
				}

				if err := LevelUp(db, &agent); err != nil {
					return fmt.Errorf(
						"level up agent %d: %w",
						result.AgentID,
						err,
					)
				}

				if err := db.CompleteJourney(result.AgentID); err != nil {
					return fmt.Errorf(
						"complete journey for agent %d: %w",
						result.AgentID,
						err,
					)
				}
			}

			continue
		}

		// Save this tick's movement first.
		err := db.ApplyAgentMovement(
			database.AgentMovementUpdate{
				AgentID:       result.AgentID,
				WorldPosition: result.Position,
				FoodOwned:     result.FoodOwned,
				PathIndex:     result.PathIndex,
				ReturningHome: result.ReturningHome,
			},
		)
		if err != nil {
			return err
		}

		// Load the updated agent for event resolution.
		agent, err := db.GetAgent(result.AgentID)
		if err != nil {
			return fmt.Errorf(
				"load agent %d for events: %w",
				result.AgentID,
				err,
			)
		}

		// 5% chance every normal travel tick.
		if err := RandomEvent(db, &agent, 5); err != nil {
			return fmt.Errorf(
				"random event for agent %d: %w",
				result.AgentID,
				err,
			)
		}

		// Guaranteed event whenever an outbound destination is reached.
		if result.ReachedDestination {
			if err := RandomEvent(db, &agent, 100); err != nil {
				return fmt.Errorf(
					"destination event for agent %d: %w",
					result.AgentID,
					err,
				)
			}
		}
	}

	return nil
}

// Tick performs one authoritative simulation update.
func Tick(db *database.Database) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	state, err := loadTickState(db)
	if err != nil {
		return err
	}

	movementResults, err := runMovementPhase(state)
	if err != nil {
		return fmt.Errorf(
			"movement phase: %w",
			err,
		)
	}

	if err := saveMovementResults(
		db,
		movementResults,
	); err != nil {
		return fmt.Errorf(
			"save movement phase: %w",
			err,
		)
	}

	if err := refreshSnapshot(db); err != nil {
		return fmt.Errorf(
			"refresh snapshot: %w",
			err,
		)
	}

	// Phase 2: detect proximity and resolve interactions.
	// Phase 3: independent events.

	return nil
}

func LevelUp(
	db *database.Database,
	agent *database.Agent,
) error {
	// We need to update the level of the agent to be equal to cubeRoot(currentExp)
	newLevel := int(math.Floor(math.Cbrt(agent.Experience)))
	if newLevel < 1 {
		newLevel = 1
	}
	if newLevel <= agent.Level {
		return nil
	}

	// We then want to update the agent's stats based on their per/level growth
	class, err := db.GetClass(agent.ClassName)
	if err != nil {
		return err
	}

	// EXAMPLE: Level 4 Warrior: 140 base HP + 8 HP/Lvl ((Level - 1) * 8 = 24)
	for agent.Level < newLevel {
		agent.MaxHealth += class.PerLevelMaxHealth
		agent.HealthRegen += class.PerLevelHealthRegen

		agent.MaxMana += class.PerLevelMaxMana
		agent.ManaRegen += class.PerLevelManaRegen

		agent.Strength += class.PerLevelStrength
		agent.Endurance += class.PerLevelEndurance
		agent.Intelligence += class.PerLevelIntelligence
		agent.Wisdom += class.PerLevelWisdom
		agent.Agility += class.PerLevelAgility
		agent.Charisma += class.PerLevelCharisma

		agent.Defense += class.PerLevelDefense
		agent.Resist += class.PerLevelResist

		agent.CriticalChance += class.PerLevelCriticalChance
		agent.CriticalDamage += class.PerLevelCriticalDamage

		agent.Speed += class.PerLevelSpeed
		agent.FoodCost += class.PerLevelFoodCost

		agent.Level++
	}
	//agent.NextLevelExp = int(math.Pow(float64(agent.Level+1), 3))

	if err := db.SaveAgentLevelUp(agent); err != nil {
		return err
	}

	return nil
}

func Refresh(
	db *database.Database,
	agent *database.Agent,
) error {
	// We need to update the level of the agent to be equal to cubeRoot(currentExp)
	agent.CurrentHealth = agent.MaxHealth
	agent.CurrentMana = agent.MaxMana

	if err := db.SaveAgent(agent); err != nil {
		return err
	}

	return nil
}
func TakeDamage(agent *database.Agent, amount float64) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}
	if amount < 0 {
		return fmt.Errorf("damage cannot be negative")
	}

	agent.CurrentHealth -= amount

	if agent.CurrentHealth < 0 {
		agent.CurrentHealth = 0
	}

	return nil
}

func GainHealth(agent *database.Agent, amount float64) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}
	if amount < 0 {
		return fmt.Errorf("health gain cannot be negative")
	}

	agent.CurrentHealth += amount

	if agent.CurrentHealth > agent.MaxHealth {
		agent.CurrentHealth = agent.MaxHealth
	}

	return nil
}
func UseMana(agent *database.Agent, amount float64) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}
	if amount < 0 {
		return fmt.Errorf("mana cost cannot be negative")
	}
	if agent.CurrentMana < amount {
		return fmt.Errorf("not enough mana")
	}

	agent.CurrentMana -= amount

	return nil
}

func GainMana(agent *database.Agent, amount float64) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}
	if amount < 0 {
		return fmt.Errorf("mana gain cannot be negative")
	}

	agent.CurrentMana += amount

	if agent.CurrentMana > agent.MaxMana {
		agent.CurrentMana = agent.MaxMana
	}

	return nil
}

func LoseMana(agent *database.Agent, amount float64) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}
	if amount < 0 {
		return fmt.Errorf("mana loss cannot be negative")
	}

	agent.CurrentMana -= amount

	if agent.CurrentMana < 0 {
		agent.CurrentMana = 0
	}

	return nil
}
func LoseFood(agent *database.Agent, amount float64) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}
	if amount < 0 {
		return fmt.Errorf("food loss cannot be negative")
	}

	agent.FoodOwned -= amount

	return nil
}

func FindFood(agent *database.Agent, amount float64) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}
	if amount < 0 {
		return fmt.Errorf("food gain cannot be negative")
	}

	agent.FoodOwned += amount

	return nil
}
func FindGold(agent *database.Agent, amount int) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}
	if amount < 0 {
		return fmt.Errorf("gold gain cannot be negative")
	}

	agent.JourneyGold += amount

	return nil
}

func LoseGold(agent *database.Agent, amount int) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}
	if amount < 0 {
		return fmt.Errorf("gold loss cannot be negative")
	}

	agent.JourneyGold -= amount

	if agent.JourneyGold < 0 {
		agent.JourneyGold = 0
	}

	return nil
}

func RollValue(base int, variance int) int {
	if variance <= 0 {
		return base
	}

	return base + rand.Intn(variance*2+1) - variance
}

func EventTable(
	db *database.Database,
	agent *database.Agent,
) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}

	name := agent.AgentName
	roll := rand.Intn(11) // 0 through 10

	switch roll {

	case 0:
		food := RollValue(2, 1)

		if err := LoseFood(agent, float64(food)); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"A Light Journey",
			fmt.Sprintf("%s stopped to smell the roses.", name),
			fmt.Sprintf(
				"Lost %d food.",
				food,
			),
		)

	case 1:
		food := RollValue(5, 2)

		if err := LoseFood(agent, float64(food)); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"A Skip and a Hop",
			fmt.Sprintf(
				"%s tripped on a rock and dropped a biscuit.",
				name,
			),
			fmt.Sprintf(
				"Lost %d food.",
				food,
			),
		)

	case 2:
		food := RollValue(10, 3)

		if err := FindFood(agent, float64(food)); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"Mushroom Picking",
			fmt.Sprintf(
				"%s found some springy mushrooms and picked them.",
				name,
			),
			fmt.Sprintf(
				"Gained %d food.",
				food,
			),
		)

	case 3:
		damage := RollValue(10, 3)

		if err := TakeDamage(agent, float64(damage)); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"Mushroom Picking",
			fmt.Sprintf(
				"%s found some suspicious mushrooms and quickly regretted eating them.",
				name,
			),
			fmt.Sprintf(
				"Poisoned. Lost %d health.",
				damage,
			),
		)

	case 4:
		food := RollValue(10, 3)

		if err := FindFood(agent, float64(food)); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"A Pleasant Stranger",
			fmt.Sprintf(
				"%s met another traveler who shared some food.",
				name,
			),
			fmt.Sprintf(
				"Gained %d food.",
				food,
			),
		)

	case 5:
		mana := RollValue(10, 2)
		damage := RollValue(4, 1)
		gold := RollValue(3, 1)

		if err := UseMana(agent, float64(mana)); err != nil {
			return err
		}

		if err := TakeDamage(agent, float64(damage)); err != nil {
			return err
		}

		if err := FindGold(agent, gold); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"Goblin Encounter",
			fmt.Sprintf(
				"A hostile goblin attacked %s!",
				name,
			),
			fmt.Sprintf(
				"Used %d mana. Took %d damage. Found %d gold.",
				mana,
				damage,
				gold,
			),
		)

	case 6:
		mana := RollValue(10, 2)

		damage1 := RollValue(4, 1)
		damage2 := RollValue(4, 1)
		totalDamage := damage1 + damage2

		gold := RollValue(6, 2)

		if err := UseMana(agent, float64(mana)); err != nil {
			return err
		}

		if err := TakeDamage(agent, float64(damage1)); err != nil {
			return err
		}

		if err := TakeDamage(agent, float64(damage2)); err != nil {
			return err
		}

		if err := FindGold(agent, gold); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"Goblin Encounter",
			fmt.Sprintf(
				"Two hostile goblins attacked %s!",
				name,
			),
			fmt.Sprintf(
				"Used %d mana. Took %d total damage. Found %d gold.",
				mana,
				totalDamage,
				gold,
			),
		)

	case 7:
		mana := RollValue(10, 2)
		damage := RollValue(4, 1)
		experience := RollValue(5, 2)
		gold := RollValue(20, 5)

		if err := UseMana(agent, float64(mana)); err != nil {
			return err
		}

		if err := TakeDamage(agent, float64(damage)); err != nil {
			return err
		}

		agent.Experience += float64(experience)

		if err := FindGold(agent, gold); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"Goblin Encounter",
			fmt.Sprintf(
				"%s defeated a goblin!",
				name,
			),
			fmt.Sprintf(
				"Used %d mana. Took %d damage. Gained %d EXP. Found %d gold.",
				mana,
				damage,
				experience,
				gold,
			),
		)

	case 8:
		damage := RollValue(12, 3)

		if err := TakeDamage(agent, float64(damage)); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"A Deadly Stranger",
			fmt.Sprintf(
				"%s got into a fight with another Hero!",
				name,
			),
			fmt.Sprintf(
				"Took %d damage.",
				damage,
			),
		)

	case 9:
		damage := RollValue(16, 4)
		experience := RollValue(20, 5)
		gold := RollValue(100, 20)

		if err := TakeDamage(agent, float64(damage)); err != nil {
			return err
		}

		agent.Experience += float64(experience)

		if err := FindGold(agent, gold); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"A Deadly Stranger",
			fmt.Sprintf(
				"%s defeated another Hero!",
				name,
			),
			fmt.Sprintf(
				"Took %d damage. Gained %d EXP. Found %d gold.",
				damage,
				experience,
				gold,
			),
		)

	case 10:
		health := RollValue(20, 5)
		mana := RollValue(20, 5)

		if err := GainHealth(agent, float64(health)); err != nil {
			return err
		}

		if err := GainMana(agent, float64(mana)); err != nil {
			return err
		}

		return ResolveEvent(
			db,
			agent,
			"Hot Springs",
			fmt.Sprintf(
				"%s found some secret hot springs!",
				name,
			),
			fmt.Sprintf(
				"Recovered %d health and %d mana.",
				health,
				mana,
			),
		)
	}

	return nil
}
func ResolveEvent(
	db *database.Database,
	agent *database.Agent,
	title string,
	description string,
	results string,
) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	if agent == nil {
		return fmt.Errorf("agent is nil")
	}

	if err := db.SaveAgent(agent); err != nil {
		return fmt.Errorf(
			"save agent %d after event: %w",
			agent.AgentID,
			err,
		)
	}

	if err := db.LogEvent(
		agent.AgentID,
		title,
		description,
		results,
	); err != nil {
		return fmt.Errorf(
			"log event for agent %d: %w",
			agent.AgentID,
			err,
		)
	}

	return nil
}

func RollEventChance(
	agent *database.Agent,
	chance float64,
) (bool, error) {
	if agent == nil {
		return false, fmt.Errorf("agent is nil")
	}

	if chance < 0 || chance > 100 {
		return false, fmt.Errorf("event chance must be between 0 and 100")
	}

	roll := rand.Intn(100) + 1 // 1 through 100

	return float64(roll) <= chance, nil
}

func RandomEvent(
	db *database.Database,
	agent *database.Agent,
	chance float64,
) error {
	triggered, err := RollEventChance(agent, chance)
	if err != nil {
		return err
	}

	if !triggered {
		return nil
	}

	if err := EventTable(db, agent); err != nil {
		return err
	}

	return nil
}
