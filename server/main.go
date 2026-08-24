package main

// Server-side database manager and simulation host.
// Admin and debugging console

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"littleadventuremap/database"
	"littleadventuremap/simulation"

	_ "github.com/lib/pq"
)

const (
	tickRate             = time.Second
	worldMin             = 0.0
	worldMax             = 100.0
	defaultDebugPassword = "password"
)

type commandInfo struct {
	Name        string
	Description string
	Shortcut    string
}

type consoleMode int

const (
	consoleAdmin consoleMode = iota
	consoleLiveAgents
)

type consoleState struct {
	mu   sync.RWMutex
	mode consoleMode
}

type simulationControl struct {
	mu     sync.RWMutex
	paused bool
}

func (c *consoleState) setMode(mode consoleMode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mode = mode
}

func (c *consoleState) isLive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mode == consoleLiveAgents
}

var consoleStateCurrent = consoleState{mode: consoleAdmin}
var simulationState simulationControl

func (s *simulationControl) isPaused() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.paused
}

func (s *simulationControl) toggle() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = !s.paused
	return s.paused
}

var adminCommands = []commandInfo{
	{Name: "help", Description: "Show available commands", Shortcut: "0"},
	{Name: "add-user", Description: "Create a test user", Shortcut: "1"},
	{Name: "print-users", Description: "View all users", Shortcut: "2"},
	{Name: "edit-user", Description: "Edit user progression values", Shortcut: "e"},
	{Name: "add-agent", Description: "Create an agent", Shortcut: "3"},
	{Name: "print-agents", Description: "View all agents", Shortcut: "4"},
	{Name: "edit-agent", Description: "Edit agent stats and attributes", Shortcut: "a"},
	{Name: "set-destination", Description: "Set one destination and deploy an agent", Shortcut: "5"},
	{Name: "print-destinations", Description: "View an agent's status and destinations", Shortcut: "6"},
	{Name: "watch-agents", Description: "Watch live agent movement on the console", Shortcut: "7"},
	{Name: "pause", Description: "Pause or resume simulation ticks", Shortcut: "p"},
	{Name: "withdraw-agent", Description: "Withdraw an agent and clear its route", Shortcut: "8"},
	{Name: "exit", Description: "Shut down the server", Shortcut: "9"},
}

func main() {
	db, err := initializeDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	ctx, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignals()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fmt.Println("Little Adventure server initialized successfully.")

	simulationDone := make(chan struct{})
	go func() {
		defer close(simulationDone)
		runSimulationRoutine(ctx, db)
	}()

	// Console input blocks, so it runs separately. This allows Ctrl+C or
	// SIGTERM to stop the server even while no command is being entered.
	go runAdminRoutine(db, cancel)

	<-ctx.Done()
	<-simulationDone

	fmt.Println("Little Adventure server stopped.")
}

func initializeDatabase() (*database.Database, error) {
	connStr := loadDatabaseURL()
	if connStr == "" {
		return nil, fmt.Errorf(
			"DATABASE_URL is not set; provide it in the environment or in an uncommitted .env file",
		)
	}

	db, err := database.NewDatabase(connStr)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return db, nil
}

func loadDatabaseURL() string {
	if connStr := strings.TrimSpace(os.Getenv("DATABASE_URL")); connStr != "" {
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

		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, value, found := strings.Cut(line, "=")
		if !found || strings.TrimSpace(key) != "DATABASE_URL" {
			continue
		}

		return strings.Trim(strings.TrimSpace(value), `"'`)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("warning: read .env: %v", err)
	}

	return ""
}

func runSimulationRoutine(
	ctx context.Context,
	db *database.Database,
) {
	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	log.Println("Simulation routine started.")
	defer log.Println("Simulation routine stopped.")

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if simulationState.isPaused() {
				continue
			}

			if err := simulation.Tick(db); err != nil {
				log.Printf("simulation tick failed: %v", err)
			}

			if consoleStateCurrent.isLive() {
				renderLiveAgentsView()
			}
		}
	}
}

func runAdminRoutine(
	db *database.Database,
	cancel context.CancelFunc,
) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Admin routine started.")
	fmt.Println("Awaiting command.")
	printCommands()

	for {
		command, err := readInput(reader, "> ")
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println()
				fmt.Println("Admin input closed. Server shutting down.")
				cancel()
				return
			}

			log.Printf("read command: %v", err)
			continue
		}

		commandName := strings.TrimSpace(command)
		if commandName == "" {
			continue
		}

		if numericCommand, err := strconv.Atoi(commandName); err == nil {
			switch numericCommand {
			case 0:
				commandName = "0: help"
			case 1:
				commandName = "1: add-user"
			case 2:
				commandName = "2: print-users"
			case 3:
				commandName = "3: add-agent"
			case 4:
				commandName = "4: print-agents"
			case 5:
				commandName = "5: set-destination"
			case 6:
				commandName = "6: print-destinations"
			case 7:
				commandName = "7: watch-agents"
			case 8:
				commandName = "8: withdraw-agent"
			case 9:
				commandName = "9: exit"
			}
		}

		switch commandName {
		case "0: help", "0", "help":
			printCommands()
		case "p", "pause":
			togglePause()
		case "1: add-user", "1", "add-user":
			addUserCommand(db, reader)
		case "2: print-users", "2", "print-users":
			printUsersCommand(db)
		case "e", "edit-user":
			editUserCommand(db, reader)
		case "3: add-agent", "3", "add-agent":
			addAgentCommand(db, reader)
		case "4: print-agents", "4", "print-agents":
			printAgentsCommand(db)
		case "a", "edit-agent":
			editAgentCommand(db, reader)
		case "5: set-destination", "5", "set-destination":
			setDestinationCommand(db, reader)
		case "6: print-destinations", "6", "print-destinations":
			printDestinationsCommand(db, reader)
		case "7: watch-agents", "7", "watch-agents":
			enterWatchMode(reader)
		case "8: withdraw-agent", "8", "withdraw-agent":
			withdrawAgentCommand(db, reader)
		case "9: exit", "9", "exit":
			fmt.Println("Admin routine stopping.")
			cancel()
			return
		default:
			fmt.Printf(
				"Unknown command %q. Enter \"help\" to see available commands.\n",
				command,
			)
		}
	}
}

func printCommands() {
	writer := newTableWriter()
	defer writer.Flush()

	fmt.Fprintln(writer, "Available commands:")
	for _, command := range adminCommands {
		fmt.Fprintf(
			writer,
			"  %s: %s\t%s\n",
			command.Shortcut,
			command.Name,
			command.Description,
		)
	}
}

func togglePause() {
	if simulationState.toggle() {
		fmt.Println("Simulation paused.")
		return
	}

	fmt.Println("Simulation resumed.")
}

func enterWatchMode(reader *bufio.Reader) {
	consoleStateCurrent.setMode(consoleLiveAgents)
	fmt.Println("Watching agents. Press Enter to return to the command prompt.")

	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		_, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("watch input: %v", err)
		}
		consoleStateCurrent.setMode(consoleAdmin)
		fmt.Println()
	}()

	<-watchDone
}

func renderLiveAgentsView() {
	clearScreen()
	fmt.Println("Little Adventure Server")
	fmt.Printf("Tick: %d\n\n", simulation.CurrentTick())

	snapshot := simulation.CurrentSnapshot()
	if len(snapshot.Agents) == 0 {
		fmt.Println("No agents found.")
		return
	}

	writer := newTableWriter()
	defer writer.Flush()

	fmt.Fprintln(
		writer,
		"ID\tName\tClass\tPosition\tHP\tMana\tEXP\tFood\tPath\tReturning\tStr\tEnd\tInt\tWis\tDex\tChr",
	)

	for _, agent := range snapshot.Agents {
		fmt.Fprintf(
			writer,
			"%d\t%s\tLv %d %s\t%s\t%.2f/%.2f\t%.2f/%.2f\t%.2f\t%.2f\t%d\t%t\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\n",
			agent.AgentID,
			agent.AgentName,
			agent.Level,
			agent.ClassName,
			formatPosition(agent.Position),
			agent.CurrentHealth,
			agent.MaxHealth,
			agent.CurrentMana,
			agent.MaxMana,
			agent.Experience,
			agent.FoodOwned,
			agent.CurrentDestinationIndex,
			agent.ReturningHome,
			agent.Strength,
			agent.Endurance,
			agent.Intelligence,
			agent.Wisdom,
			agent.Agility,
			agent.Charisma,
		)
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func addUserCommand(
	db *database.Database,
	reader *bufio.Reader,
) {
	username, err := readRequiredInput(reader, "Username: ")
	if err != nil {
		fmt.Printf("Could not read username: %v\n", err)
		return
	}

	userID, err := db.CreateUser(username, defaultDebugPassword)
	if err != nil {
		fmt.Printf("Could not create user: %v\n", err)
		return
	}

	fmt.Printf(
		"Created user %q with owner ID %d.\n",
		username,
		userID,
	)
}

func printUsersCommand(db *database.Database) {
	users, err := db.GetUsers()
	if err != nil {
		fmt.Printf("Could not load users: %v\n", err)
		return
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return
	}

	writer := newTableWriter()
	defer writer.Flush()

	fmt.Fprintln(writer, "Users:")
	fmt.Fprintln(writer, "ID\tUsername\tMoney\tMax Agents\tGuild Level\tGuild EXP\tGuild Next Level EXP")
	for _, user := range users {
		fmt.Fprintf(writer, "%d\t%s\t%d\t%d\t%d\t%.2f\t%d\n",
			user.UserID, user.Username, user.Money, user.MaxAgents,
			user.GuildLevel, user.GuildExperience, user.GuildNextLevelExp)
	}
}

func editUserCommand(db *database.Database, reader *bufio.Reader) {
	userID, err := readInt64Input(reader, "User ID: ")
	if err != nil {
		fmt.Printf("Invalid user ID: %v\n", err)
		return
	}

	money, err := readInt64Input(reader, "Money: ")
	if err != nil {
		fmt.Printf("Invalid money: %v\n", err)
		return
	}
	maxAgents, err := readInt64Input(reader, "Max agents: ")
	if err != nil {
		fmt.Printf("Invalid max agents: %v\n", err)
		return
	}
	guildLevel, err := readInt64Input(reader, "Guild level: ")
	if err != nil {
		fmt.Printf("Invalid guild level: %v\n", err)
		return
	}
	guildExperience, err := readFloatInput(reader, "Guild EXP: ")
	if err != nil {
		fmt.Printf("Invalid guild EXP: %v\n", err)
		return
	}

	if money < 0 || maxAgents < 0 || guildLevel < 1 || guildExperience < 0 {
		fmt.Println("Money, max agents, and guild EXP cannot be negative; guild level must be at least 1.")
		return
	}

	if err := db.UpdateUserProgress(
		userID,
		int(money),
		int(maxAgents),
		int(guildLevel),
		guildExperience,
	); err != nil {
		fmt.Printf("Could not update user: %v\n", err)
		return
	}
	fmt.Printf("Updated user %d. Guild next level EXP is %d.\n", userID, levelExperienceTarget(int(guildLevel)))
}

func addAgentCommand(
	db *database.Database,
	reader *bufio.Reader,
) {
	ownerID, err := readInt64Input(reader, "Owner ID: ")
	if err != nil {
		fmt.Printf("Invalid owner ID: %v\n", err)
		return
	}

	agentName, err := readRequiredInput(reader, "Agent name: ")
	if err != nil {
		fmt.Printf("Could not read agent name: %v\n", err)
		return
	}

	agentID, err := db.CreateAgent(ownerID, agentName)
	if err != nil {
		fmt.Printf("Could not create agent: %v\n", err)
		return
	}

	fmt.Printf(
		"Created agent %q with agent ID %d for owner %d.\n",
		agentName,
		agentID,
		ownerID,
	)
}

func printAgentsCommand(db *database.Database) {
	snapshot, err := simulation.BuildSnapshot(db)
	if err != nil {
		fmt.Printf("Could not build agent snapshot: %v\n", err)
		return
	}

	if len(snapshot.Agents) == 0 {
		fmt.Println("No agents found.")
		return
	}

	writer := newTableWriter()
	defer writer.Flush()

	fmt.Fprintln(
		writer,
		"ID\tOwner\tName\tPosition\tHP\tMana\tLevel\tEXP\tNext Level EXP\tFood\tPath Index\tReturning\tStrength\tEndurance\tIntelligence\tWisdom\tAgility\tCharisma\tSpeed\tFood Cost\tDeployed",
	)

	for _, agent := range snapshot.Agents {
		fmt.Fprintf(
			writer,
			"%d\t%d\t%s\t%s\t%.2f/%d\t%.2f/%d\t%d\t%.2f\t%d\t%.2f\t%d\t%t\t%d\t%d\t%d\t%d\t%d\t%d\t%.2f\t%.2f\t%t\n",
			agent.AgentID,
			agent.OwnerID,
			agent.AgentName,
			formatPosition(agent.Position),
			agent.CurrentHealth,
			agent.MaxHealth,
			agent.CurrentMana,
			agent.MaxMana,
			agent.Level,
			agent.Experience,
			agent.NextLevelExp,
			agent.FoodOwned,
			agent.PathIndex,
			agent.ReturningHome,
			agent.Strength,
			agent.Endurance,
			agent.Intelligence,
			agent.Wisdom,
			agent.Agility,
			agent.Charisma,
			agent.Speed,
			agent.FoodCost,
			agent.Deployed,
		)
	}
}

func editAgentCommand(db *database.Database, reader *bufio.Reader) {
	agentID, err := readInt64Input(reader, "Agent ID: ")
	if err != nil {
		fmt.Printf("Invalid agent ID: %v\n", err)
		return
	}

	currentHealth, err := readFloatInput(reader, "Current health: ")
	if err != nil {
		fmt.Printf("Invalid current health: %v\n", err)
		return
	}

	maxHealth, err := readFloatInput(reader, "Max health: ")
	if err != nil {
		fmt.Printf("Invalid max health: %v\n", err)
		return
	}

	currentMana, err := readFloatInput(reader, "Current mana: ")
	if err != nil {
		fmt.Printf("Invalid current mana: %v\n", err)
		return
	}

	maxMana, err := readFloatInput(reader, "Max mana: ")
	if err != nil {
		fmt.Printf("Invalid max mana: %v\n", err)
		return
	}

	attributes := make([]float64, 6)
	attributeNames := []string{
		"Strength",
		"Endurance",
		"Intelligence",
		"Wisdom",
		"Agility",
		"Charisma",
	}

	for index, name := range attributeNames {
		value, err := readFloatInput(reader, name+": ")
		if err != nil {
			fmt.Printf("Invalid %s: %v\n", name, err)
			return
		}

		attributes[index] = value
	}

	experience, err := readFloatInput(reader, "Experience: ")
	if err != nil {
		fmt.Printf("Invalid experience: %v\n", err)
		return
	}

	level, err := readInt64Input(reader, "Level: ")
	if err != nil {
		fmt.Printf("Invalid level: %v\n", err)
		return
	}

	if maxHealth < 0 ||
		maxMana < 0 ||
		currentHealth < 0 ||
		currentMana < 0 ||
		currentHealth > maxHealth ||
		currentMana > maxMana ||
		experience < 0 ||
		level < 1 {
		fmt.Println(
			"Health/mana must be within their maximums, EXP cannot be negative, and level must be at least 1.",
		)
		return
	}

	if err := db.UpdateAgentStats(
		agentID,
		currentHealth,
		maxHealth,
		currentMana,
		maxMana,
		attributes[0],
		attributes[1],
		attributes[2],
		attributes[3],
		attributes[4],
		attributes[5],
		experience,
		int(level),
	); err != nil {
		fmt.Printf("Could not update agent: %v\n", err)
		return
	}

	fmt.Printf(
		"Updated agent %d. Next level EXP is %d.\n",
		agentID,
		levelExperienceTarget(int(level)),
	)
}

func setDestinationCommand(
	db *database.Database,
	reader *bufio.Reader,
) {
	ownerID, agentID, ok := selectOwnedAgent(db, reader)
	if !ok {
		return
	}

	agent, err := db.GetAgentByOwner(ownerID, agentID)
	if err != nil {
		fmt.Printf("Could not load agent: %v\n", err)
		return
	}

	destination, ok := readWorldPosition(reader, "Destination")
	if !ok {
		fmt.Println("Destination canceled.")
		return
	}

	if agent.Deployed {
		if err := db.AddDestination(agentID, destination); err != nil {
			fmt.Printf("Could not add destination: %v\n", err)
			return
		}

		fmt.Printf(
			"Added destination (%.2f, %.2f) to agent %d.\n",
			destination[0],
			destination[1],
			agentID,
		)
		return
	}

	homePosition, ok := readStartingPosition(reader)
	if !ok {
		fmt.Println("Deployment canceled.")
		return
	}

	if err := db.SetDestinationAndDeploy(
		ownerID,
		agentID,
		homePosition,
		destination,
	); err != nil {
		fmt.Printf("Could not deploy agent: %v\n", err)
		return
	}

	deployedAgent, err := db.GetAgentByOwner(ownerID, agentID)
	if err != nil {
		fmt.Printf(
			"Agent %d deployed from (%.2f, %.2f) toward (%.2f, %.2f).\n",
			agentID, homePosition[0], homePosition[1], destination[0], destination[1],
		)
		return
	}

	fmt.Printf(
		"Agent %d deployed from (%.2f, %.2f) toward (%.2f, %.2f) with %.2f food.\n",
		agentID, homePosition[0], homePosition[1], destination[0], destination[1],
		deployedAgent.FoodOwned,
	)
}

func printDestinationsCommand(
	db *database.Database,
	reader *bufio.Reader,
) {
	ownerID, agentID, ok := selectOwnedAgent(db, reader)
	if !ok {
		return
	}

	agent, err := db.GetAgentByOwner(ownerID, agentID)
	if err != nil {
		fmt.Printf("Could not load agent: %v\n", err)
		return
	}

	destinations, err := db.GetDestinationsByAgent(agentID)
	if err != nil {
		fmt.Printf("Could not load destinations: %v\n", err)
		return
	}

	fmt.Println()
	printAgentStatus(agent, destinations)
	printAgentDestinations(agentID, destinations)
}

func withdrawAgentCommand(
	db *database.Database,
	reader *bufio.Reader,
) {
	agentID, err := readInt64Input(reader, "Agent ID: ")
	if err != nil {
		fmt.Printf("Invalid agent ID: %v\n", err)
		return
	}

	if err := db.WithdrawAgent(agentID); err != nil {
		fmt.Printf("Could not withdraw agent: %v\n", err)
		return
	}

	fmt.Printf(
		"Agent %d was withdrawn and its destinations were cleared.\n",
		agentID,
	)
}

func selectOwnedAgent(
	db *database.Database,
	reader *bufio.Reader,
) (int64, int64, bool) {
	ownerID, err := readInt64Input(reader, "Owner ID: ")
	if err != nil {
		fmt.Printf("Invalid owner ID: %v\n", err)
		return 0, 0, false
	}

	agents, err := db.GetAgentsByOwner(ownerID)
	if err != nil {
		fmt.Printf("Could not load agents: %v\n", err)
		return 0, 0, false
	}

	if len(agents) == 0 {
		fmt.Printf("Owner %d has no agents.\n", ownerID)
		return 0, 0, false
	}

	printAgentSelectionTable(agents)

	agentID, err := readInt64Input(reader, "Agent ID: ")
	if err != nil {
		fmt.Printf("Invalid agent ID: %v\n", err)
		return 0, 0, false
	}

	for _, agent := range agents {
		if agent.AgentID == agentID {
			return ownerID, agentID, true
		}
	}

	fmt.Printf(
		"Agent %d does not belong to owner %d.\n",
		agentID,
		ownerID,
	)
	return 0, 0, false
}

func printAgentSelectionTable(agents []database.Agent) {
	writer := newTableWriter()
	defer writer.Flush()

	fmt.Fprintln(
		writer,
		"ID\tName\tSpeed\tFood Cost\tFood Owned\tDeployed",
	)

	for _, agent := range agents {
		fmt.Fprintf(
			writer,
			"%d\t%s\t%.2f\t%.2f\t%.2f\t%t\n",
			agent.AgentID,
			agent.AgentName,
			agent.Speed,
			agent.FoodCost,
			agent.FoodOwned,
			agent.Deployed,
		)
	}
}

func printAgentStatus(agent database.Agent, destinations []database.Destination) {
	writer := newTableWriter()
	defer writer.Flush()

	heading := "Idle"
	var currentDestination *[2]float64

	if agent.Deployed {
		if agent.ReturningHome {
			heading = "Returning Home"
		} else {
			heading = "Outbound"
		}

		if len(destinations) > 0 {
			if agent.PathIndex >= 0 && agent.PathIndex < len(destinations) {
				position := destinations[agent.PathIndex].WorldPosition
				currentDestination = &position
			} else if agent.ReturningHome && agent.HomePosition != nil {
				position := [2]float64{(*agent.HomePosition)[0], (*agent.HomePosition)[1]}
				currentDestination = &position
			}
		}
	}

	fmt.Fprintln(writer, "Agent Status:")
	fmt.Fprintln(writer, "Field\tValue")
	fmt.Fprintf(writer, "Agent ID\t%d\n", agent.AgentID)
	fmt.Fprintf(writer, "Owner ID\t%d\n", agent.OwnerID)
	fmt.Fprintf(writer, "Name\t%s\n", agent.AgentName)
	fmt.Fprintf(writer, "Speed\t%.2f\n", agent.Speed)
	fmt.Fprintf(writer, "Food Cost\t%.2f\n", agent.FoodCost)
	fmt.Fprintf(writer, "Food Owned\t%.2f\n", agent.FoodOwned)
	fmt.Fprintf(writer, "Current Health\t%.2f\n", agent.CurrentHealth)
	fmt.Fprintf(writer, "Max Health\t%d\n", agent.MaxHealth)
	fmt.Fprintf(writer, "Current Mana\t%.2f\n", agent.CurrentMana)
	fmt.Fprintf(writer, "Max Mana\t%d\n", agent.MaxMana)
	fmt.Fprintf(writer, "Level\t%d\n", agent.Level)
	fmt.Fprintf(writer, "Experience\t%.2f\n", agent.Experience)
	fmt.Fprintf(writer, "Next Level EXP\t%d\n", agent.NextLevelExp)
	fmt.Fprintf(writer, "Strength\t%d\n", agent.Strength)
	fmt.Fprintf(writer, "Endurance\t%d\n", agent.Endurance)
	fmt.Fprintf(writer, "Intelligence\t%d\n", agent.Intelligence)
	fmt.Fprintf(writer, "Wisdom\t%d\n", agent.Wisdom)
	fmt.Fprintf(writer, "Agility\t%d\n", agent.Agility)
	fmt.Fprintf(writer, "Charisma\t%d\n", agent.Charisma)
	fmt.Fprintf(writer, "Deployed\t%t\n", agent.Deployed)
	fmt.Fprintf(writer, "Destination Count\t%d\n", len(destinations))
	fmt.Fprintf(writer, "Current Destination Index\t%d\n", agent.PathIndex)
	fmt.Fprintf(writer, "Current Destination\t%s\n", formatPosition(currentDestination))
	fmt.Fprintf(writer, "Heading\t%s\n", heading)
	fmt.Fprintf(writer, "World Position\t%s\n", formatPosition(agent.WorldPosition))
	fmt.Fprintf(writer, "Home Position\t%s\n", formatPosition(agent.HomePosition))
}

func printAgentDestinations(
	agentID int64,
	destinations []database.Destination,
) {
	if len(destinations) == 0 {
		fmt.Printf("\nAgent %d has no destinations.\n", agentID)
		return
	}

	writer := newTableWriter()
	defer writer.Flush()

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Destinations:")
	fmt.Fprintln(writer, "ID\tOrder\tName\tX\tY")

	for _, destination := range destinations {
		fmt.Fprintf(
			writer,
			"%d\t%d\t%s\t%.2f\t%.2f\n",
			destination.DestinationID,
			destination.DestinationOrder,
			destination.DestinationName,
			destination.WorldPosition[0],
			destination.WorldPosition[1],
		)
	}
}

func readStartingPosition(
	reader *bufio.Reader,
) ([2]float64, bool) {
	fmt.Println("Starting locations:")
	fmt.Printf("  A : (%.0f, %.0f)\n", worldMin, worldMin)
	fmt.Printf("  B : (%.0f, %.0f)\n", worldMax, worldMin)
	fmt.Printf("  C : (%.0f, %.0f)\n", worldMin, worldMax)
	fmt.Printf("  D : (%.0f, %.0f)\n", worldMax, worldMax)

	choice, err := readInput(
		reader,
		"Select starting location [A/B/C/D]: ",
	)
	if err != nil {
		fmt.Printf("Could not read starting location: %v\n", err)
		return [2]float64{}, false
	}

	switch strings.ToUpper(choice) {
	case "A":
		return [2]float64{worldMin, worldMin}, true
	case "B":
		return [2]float64{worldMax, worldMin}, true
	case "C":
		return [2]float64{worldMin, worldMax}, true
	case "D":
		return [2]float64{worldMax, worldMax}, true
	default:
		return [2]float64{}, false
	}
}

func readWorldPosition(
	reader *bufio.Reader,
	label string,
) ([2]float64, bool) {
	x, err := readFloatInput(
		reader,
		fmt.Sprintf("%s X (%.0f-%.0f): ", label, worldMin, worldMax),
	)
	if err != nil {
		fmt.Printf("Invalid X coordinate: %v\n", err)
		return [2]float64{}, false
	}

	y, err := readFloatInput(
		reader,
		fmt.Sprintf("%s Y (%.0f-%.0f): ", label, worldMin, worldMax),
	)
	if err != nil {
		fmt.Printf("Invalid Y coordinate: %v\n", err)
		return [2]float64{}, false
	}

	if !isWithinWorld(x) || !isWithinWorld(y) {
		fmt.Printf(
			"Position is outside the world bounds (%.0f-%.0f).\n",
			worldMin,
			worldMax,
		)
		return [2]float64{}, false
	}

	return [2]float64{x, y}, true
}

func readInput(
	reader *bufio.Reader,
	prompt string,
) (string, error) {
	fmt.Print(prompt)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

func readRequiredInput(
	reader *bufio.Reader,
	prompt string,
) (string, error) {
	value, err := readInput(reader, prompt)
	if err != nil {
		return "", err
	}

	if value == "" {
		return "", fmt.Errorf("a value is required")
	}

	return value, nil
}

func readInt64Input(
	reader *bufio.Reader,
	prompt string,
) (int64, error) {
	input, err := readRequiredInput(reader, prompt)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid integer", input)
	}

	return value, nil
}

func readFloatInput(
	reader *bufio.Reader,
	prompt string,
) (float64, error) {
	input, err := readRequiredInput(reader, prompt)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid number", input)
	}

	return value, nil
}

func levelExperienceTarget(level int) int {
	if level < 1 {
		return 0
	}
	return level * level * level
}

func newTableWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(
		os.Stdout,
		0,
		4,
		2,
		' ',
		0,
	)
}

func isWithinWorld(value float64) bool {
	return value >= worldMin && value <= worldMax
}

func formatPosition(position *[2]float64) string {
	if position == nil {
		return "-"
	}

	return fmt.Sprintf(
		"(%.2f, %.2f)",
		position[0],
		position[1],
	)
}
