package main

// Client side interface
// Allows client to interact with game database

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"littleadventuremap/database"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	connStr := loadDatabaseURL()
	if connStr == "" {
		log.Fatal(
			"DATABASE_URL is not set; provide it in the environment or in an uncommitted .env file",
		)
	}

	db, err := database.NewDatabase(connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	userID, err := login(reader, db)
	if err != nil {
		if errors.Is(err, io.EOF) {
			fmt.Println("\nInput closed.")
			return
		}

		log.Fatal(err)
	}

	fmt.Println("\nLogin successful.")

	if err := db.PrintUserSummary(userID); err != nil {
		log.Printf("Could not load guild information: %v", err)
	}

	runClientLoop(reader, db, userID)
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

		line = strings.TrimSpace(
			strings.TrimPrefix(line, "export "),
		)

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		if strings.TrimSpace(key) != "DATABASE_URL" {
			continue
		}

		return strings.Trim(
			strings.TrimSpace(value),
			`"'`,
		)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("warning: read .env: %v", err)
	}

	return ""
}

func login(
	reader *bufio.Reader,
	db *database.Database,
) (int64, error) {
	for {
		fmt.Println("\nWelcome to Elderwood Mini: Please Log In")

		username, err := readRequiredInput(reader, "Username: ")
		if err != nil {
			return 0, err
		}

		password, err := readPassword("Password: ")
		if err != nil {
			return 0, err
		}

		if password == "" {
			fmt.Println("Password cannot be empty.")
			continue
		}

		user, err := db.GetUserByUsername(username)

		// Existing account.
		if err == nil {
			if err := bcrypt.CompareHashAndPassword(
				[]byte(user.PasswordHash),
				[]byte(password),
			); err != nil {
				fmt.Println("Incorrect username or password.")
				continue
			}

			return user.UserID, nil
		}

		// Real database error.
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("look up user: %w", err)
		}

		// Account does not exist.
		answer, err := readInput(
			reader,
			"Account does not exist. Create it? [y/N]: ",
		)
		if err != nil {
			return 0, err
		}

		answer = strings.ToLower(answer)

		if answer != "y" && answer != "yes" {
			continue
		}

		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return 0, fmt.Errorf("hash password: %w", err)
		}

		userID, err := db.CreateUser(
			username,
			string(passwordHash),
		)
		if err != nil {
			return 0, fmt.Errorf("create user: %w", err)
		}

		fmt.Println("Account created with 1000 starting Gold.")

		return userID, nil
	}
}

func readInput(
	reader *bufio.Reader,
	prompt string,
) (string, error) {
	fmt.Print(prompt)

	input, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && len(input) > 0 {
			return strings.TrimSpace(input), nil
		}

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

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	passwordBytes, err := term.ReadPassword(
		int(os.Stdin.Fd()),
	)
	fmt.Println()

	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}

	return strings.TrimSpace(string(passwordBytes)), nil
}

func runClientLoop(
	reader *bufio.Reader,
	db *database.Database,
	userID int64,
) {
	for {
		printCommandMenu()

		command, err := readInput(reader, "> ")
		if err != nil {
			fmt.Printf("Could not read command: %v\n", err)
			return
		}

		if !handleCommand(reader, db, userID, command) {
			return
		}
	}
}

func printCommandMenu() {
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("1: Hire a Hero")
	fmt.Println("2: Deploy a Hero")
	fmt.Println("3: Manage Heroes")
	fmt.Println("4: View Inventory")
	fmt.Println("5: Market")
	fmt.Println("6: Library")
	fmt.Println("7: Quit")
}

func handleCommand(
	reader *bufio.Reader,
	db *database.Database,
	userID int64,
	command string,
) bool {
	switch strings.ToLower(command) {
	case "1", "hire a hero", "hire":
		hireHeroMenu(reader, db, userID)

	case "2", "deploy a hero", "deploy":
		deployHeroMenu(reader, db, userID)

	case "3", "manage heroes", "heroes":
		manageHeroesMenu(reader, db, userID)

	case "4", "view inventory", "inventory":
		fmt.Println("View Inventory selected.")

	case "5", "market":
		fmt.Println("Market selected.")

	case "6", "library":
		fmt.Println("Library selected.")

	case "7", "quit", "exit":
		fmt.Println("Goodbye.")
		return false

	default:
		fmt.Printf("Unknown command %q.\n", command)
	}

	return true
}

func manageHeroesMenu(
	reader *bufio.Reader,
	db *database.Database,
	userID int64,
) {
	agents, err := db.GetAgentsByOwner(userID)
	if err != nil {
		fmt.Printf("Could not load your Heroes: %v\n", err)
		return
	}

	if len(agents) == 0 {
		fmt.Println()
		fmt.Println("You do not own any Heroes.")
		return
	}

	fmt.Println()
	fmt.Println("Please select a Hero to inspect:")
	fmt.Println()

	for index, agent := range agents {
		status := "At Home"
		if agent.Deployed {
			status = "Deployed"
		}

		fmt.Printf(
			"%d. %s | Level %d %s | Status: %s\n",
			index+1,
			agent.AgentName,
			agent.Level,
			agent.ClassName,
			status,
		)
	}

	fmt.Println()
	fmt.Println("0) Return to Home Menu")

	choice, err := readInput(reader, "> ")
	if err != nil {
		fmt.Printf("Could not read Hero selection: %v\n", err)
		return
	}

	choice = strings.TrimSpace(choice)

	if choice == "0" {
		return
	}

	selection, err := strconv.Atoi(choice)
	if err != nil || selection < 1 || selection > len(agents) {
		fmt.Println("Invalid Hero selection.")
		return
	}

	manageHeroMenu(
		reader,
		db,
		userID,
		agents[selection-1],
	)
}
func manageHeroMenu(
	reader *bufio.Reader,
	db *database.Database,
	userID int64,
	agent database.Agent,
) {
	for {
		printHeroDetails(agent)

		fmt.Println()
		fmt.Println("1. Change Equipment")
		fmt.Println("2. Change Inventory")
		fmt.Println("3. Dismiss")
		fmt.Println()
		fmt.Println("0) Return to Home Menu")

		choice, err := readInput(reader, "> ")
		if err != nil {
			fmt.Printf("Could not read selection: %v\n", err)
			return
		}

		switch strings.TrimSpace(choice) {
		case "1":
			fmt.Println("Changing Equipment is not implemented yet.")

		case "2":
			fmt.Println("Changing Inventory is not implemented yet.")

		case "3":
			dismissHero(
				reader,
				db,
				userID,
				agent,
			)
			return

		case "0":
			return

		default:
			fmt.Println("Invalid selection.")
		}
	}
}

func printHeroDetails(agent database.Agent) {
	status := "At Home"
	if agent.Deployed {
		status = "Deployed"
	}

	fmt.Println()
	fmt.Printf(
		"%s | Level %d %s | Status: %s\n",
		agent.AgentName,
		agent.Level,
		agent.ClassName,
		status,
	)
	fmt.Println("--------------------------------------------------")

	fmt.Printf(
		"Health: %.2f / %.2f | Mana: %.2f / %.2f\n",
		agent.CurrentHealth,
		agent.MaxHealth,
		agent.CurrentMana,
		agent.MaxMana,
	)

	expToNext := math.Max(
		float64(agent.NextLevelExp)-agent.Experience,
		0,
	)

	fmt.Printf(
		"Level %d %s | %.2f EXP to next\n",
		agent.Level,
		agent.ClassName,
		expToNext,
	)

	fmt.Printf(
		"Food Cost: %.2f | Speed: %.2f\n",
		agent.FoodCost,
		agent.Speed,
	)

	fmt.Println()
	fmt.Println("Attributes")
	fmt.Printf("Strength:     %.2f\n", agent.Strength)
	fmt.Printf("Endurance:    %.2f\n", agent.Endurance)
	fmt.Printf("Intelligence: %.2f\n", agent.Intelligence)
	fmt.Printf("Wisdom:       %.2f\n", agent.Wisdom)
	fmt.Printf("Agility:      %.2f\n", agent.Agility)
	fmt.Printf("Charisma:     %.2f\n", agent.Charisma)
}

func dismissHero(
	reader *bufio.Reader,
	db *database.Database,
	userID int64,
	agent database.Agent,
) {
	if agent.Deployed {
		fmt.Printf(
			"%s cannot be dismissed while deployed.\n",
			agent.AgentName,
		)
		return
	}

	fmt.Println()
	fmt.Printf(
		"Dismiss %s the %s?\n",
		agent.AgentName,
		agent.ClassName,
	)

	fmt.Println(
		"This permanently removes the Hero from your Guild.",
	)

	answer, err := readInput(
		reader,
		"Confirm Dismissal [y/N]: ",
	)
	if err != nil {
		fmt.Printf("Could not read confirmation: %v\n", err)
		return
	}

	if !isYes(answer) {
		fmt.Println("Dismissal canceled.")
		return
	}

	if err := db.ReleaseHero(
		userID,
		agent.AgentID,
	); err != nil {
		fmt.Printf(
			"Could not dismiss Hero: %v\n",
			err,
		)
		return
	}

	fmt.Printf(
		"%s has been dismissed from your Guild.\n",
		agent.AgentName,
	)
}

func hireHeroMenu(
	reader *bufio.Reader,
	db *database.Database,
	userID int64,
) {
	hireState, err := db.GetHireState(userID)
	if err != nil {
		fmt.Printf("Could not check Hero capacity: %v\n", err)
		return
	}

	if hireState.CurrentAgents >= hireState.MaxAgents {
		fmt.Printf(
			"\nHero capacity reached: %d/%d\n",
			hireState.CurrentAgents,
			hireState.MaxAgents,
		)
		return
	}

	const hireCost = 600

	fmt.Println()
	fmt.Println("Hire a Hero")
	fmt.Println("1: Warrior - Strong, well balanced hero. Requires slightly more food.")
	fmt.Println("2: Rogue - Moves across the map quickly.")
	fmt.Println("3: Ranger - Moves across the map efficiently, requiring less food.")
	fmt.Println("4: Wizard - Generates mana rapidly.")
	fmt.Println("5: Cleric - Regenerates health quickly.")
	fmt.Println("0: Cancel")

	choice, err := readInput(reader, "> ")
	if err != nil {
		fmt.Printf("Could not read Hero selection: %v\n", err)
		return
	}

	var className string

	switch strings.TrimSpace(choice) {
	case "1":
		className = "Warrior"
	case "2":
		className = "Rogue"
	case "3":
		className = "Ranger"
	case "4":
		className = "Wizard"
	case "5":
		className = "Cleric"
	case "0":
		return
	default:
		fmt.Println("Invalid Hero selection.")
		return
	}

	class, err := db.GetClass(className)
	if err != nil {
		fmt.Printf("Could not load %s: %v\n", className, err)
		return
	}

	printClassPreview(class)

	prompt := fmt.Sprintf(
		"Hire this Hero for %d Gold? (Current money %d Gold) [y/N]: ",
		hireCost,
		hireState.Gold,
	)

	answer, err := readInput(
		reader,
		prompt,
	)
	if err != nil {
		fmt.Printf("Could not read response: %v\n", err)
		return
	}

	if !isYes(answer) {
		return
	}

	if hireState.Gold < hireCost {
		fmt.Printf(
			"You need %d Gold to hire this Hero, but have %d.\n",
			hireCost,
			hireState.Gold,
		)
		return
	}

	heroName, err := readInput(
		reader,
		fmt.Sprintf("Name your %s [default: %s]: ", className, className),
	)
	if err != nil {
		fmt.Printf("Could not read Hero name: %v\n", err)
		return
	}

	heroName = strings.TrimSpace(heroName)
	if heroName == "" {
		heroName = className
	}

	agentID, err := db.HireHero(userID, heroName, className)
	if err != nil {
		fmt.Printf("Could not hire Hero: %v\n", err)
		return
	}

	fmt.Printf(
		"%s has joined your Guild! [Hero ID %d]\n",
		heroName,
		agentID,
	)
}

func printClassPreview(class database.Class) {
	fmt.Println()
	fmt.Printf("%s | Level 1\n", class.ClassName)
	fmt.Println("--------------------------------------------------")

	fmt.Printf(
		"Food Cost: %.2f | Speed: %.2f\n",
		class.BaseFoodCost,
		class.BaseSpeed,
	)

	fmt.Printf("Weapon: %s\n", class.BaseWeapon)
	fmt.Printf("Armor:  %s\n", class.BaseArmor)

	fmt.Println()
	fmt.Println("Resources")
	fmt.Printf(
		"Health:       %.2f + %.2f / level\n",
		class.BaseMaxHealth,
		class.PerLevelMaxHealth,
	)
	fmt.Printf(
		"Health Regen: %.3f + %.3f / level\n",
		class.BaseHealthRegen,
		class.PerLevelHealthRegen,
	)
	fmt.Printf(
		"Mana:         %.2f + %.2f / level\n",
		class.BaseMaxMana,
		class.PerLevelMaxMana,
	)
	fmt.Printf(
		"Mana Regen:   %.3f + %.3f / level\n",
		class.BaseManaRegen,
		class.PerLevelManaRegen,
	)

	fmt.Println()
	fmt.Println("Attributes")
	fmt.Printf(
		"Strength:     %.2f + %.2f / level\n",
		class.BaseStrength,
		class.PerLevelStrength,
	)
	fmt.Printf(
		"Endurance:    %.2f + %.2f / level\n",
		class.BaseEndurance,
		class.PerLevelEndurance,
	)
	fmt.Printf(
		"Intelligence: %.2f + %.2f / level\n",
		class.BaseIntelligence,
		class.PerLevelIntelligence,
	)
	fmt.Printf(
		"Wisdom:       %.2f + %.2f / level\n",
		class.BaseWisdom,
		class.PerLevelWisdom,
	)
	fmt.Printf(
		"Agility:      %.2f + %.2f / level\n",
		class.BaseAgility,
		class.PerLevelAgility,
	)
	fmt.Printf(
		"Charisma:     %.2f + %.2f / level\n",
		class.BaseCharisma,
		class.PerLevelCharisma,
	)

	fmt.Println()
	fmt.Println("Combat")
	fmt.Printf(
		"Defense:      %.2f + %.2f / level\n",
		class.BaseDefense,
		class.PerLevelDefense,
	)
	fmt.Printf(
		"Resist:       %.2f + %.2f / level\n",
		class.BaseResist,
		class.PerLevelResist,
	)
	fmt.Printf(
		"Critical:     %.2f + %.2f / level\n",
		class.BaseCriticalChance,
		class.PerLevelCriticalChance,
	)
	fmt.Printf(
		"Critical Dmg: %.2f + %.2f / level\n",
		class.BaseCriticalDamage,
		class.PerLevelCriticalDamage,
	)

	fmt.Println()
}

func deployHeroMenu(
	reader *bufio.Reader,
	db *database.Database,
	userID int64,
) {
	agents, err := db.GetAgentsByOwner(userID)
	if err != nil {
		fmt.Printf("Could not load your Heroes: %v\n", err)
		return
	}

	if len(agents) == 0 {
		fmt.Println()
		fmt.Println("You do not own any Heroes.")
		fmt.Println("Hire a Hero before attempting to deploy.")
		return
	}

	fmt.Println()
	fmt.Println("Deploy a Hero")

	for index, agent := range agents {
		status := "At Home"
		if agent.Deployed {
			status = "Deployed"
		}

		fmt.Printf(
			"%d. %s | Level %d %s | Status: %s\n",
			index+1,
			agent.AgentName,
			agent.Level,
			agent.ClassName,
			status,
		)
	}

	fmt.Println("0. Cancel")

	choice, err := readInput(reader, "> ")
	if err != nil {
		fmt.Printf("Could not read Hero selection: %v\n", err)
		return
	}

	if strings.TrimSpace(choice) == "0" {
		return
	}

	selection, err := strconv.Atoi(strings.TrimSpace(choice))
	if err != nil || selection < 1 || selection > len(agents) {
		fmt.Println("Invalid Hero selection.")
		return
	}

	agent := agents[selection-1]
	if agent.Deployed {
		fmt.Printf("%s is already deployed.\n", agent.AgentName)
		return
	}

	homePosition, ok := readStartingPosition(reader)
	if !ok {
		fmt.Println("Deployment canceled.")
		return
	}

	destinations := readDeploymentDestinations(reader)
	if len(destinations) == 0 {
		fmt.Println("No destinations entered. Deployment canceled.")
		return
	}

	totalDistance := calculateRouteDistance(homePosition, destinations)
	requiredFood := math.Ceil(totalDistance * agent.FoodCost)
	const freeFood = 100.0

	fmt.Println()
	fmt.Printf("Total Distance: %.2f\n", totalDistance)
	fmt.Printf("Food Required: %.0f\n", requiredFood)
	fmt.Printf("Deployment Food: %.0f\n", freeFood)

	additionalFood := openFoodStoreMenu(
		reader,
		agent,
		requiredFood,
	)

	totalFood := freeFood + float64(additionalFood)

	fmt.Println()
	fmt.Printf(
		"%s | Level %d %s | %.0f Deployment Food\n",
		agent.AgentName,
		agent.Level,
		agent.ClassName,
		totalFood,
	)
	fmt.Printf(
		"%d Destinations for a total of %.2f distance\n",
		len(destinations),
		totalDistance,
	)

	answer, err := readInput(reader, "Confirm Deployment [y/N]: ")
	if err != nil {
		fmt.Printf("Could not read confirmation: %v\n", err)
		return
	}

	if !isYes(answer) {
		return
	}

	if err := db.DeployAgentRoute(
		userID,
		agent.AgentID,
		homePosition,
		destinations,
		additionalFood,
	); err != nil {
		fmt.Printf("Could not deploy Hero: %v\n", err)
		return
	}

	fmt.Printf("%s has begun their adventure.\n", agent.AgentName)
}
func openFoodStoreMenu(
	reader *bufio.Reader,
	agent database.Agent,
	requiredFood float64,
) int {
	const freeFood = 100.0

	foodDifference := math.Max(requiredFood-freeFood, 0)

	fmt.Println()
	fmt.Println("Purchase Additional Food")
	fmt.Println("----------------------------------------")
	fmt.Printf("Hero: %s\n", agent.AgentName)
	fmt.Printf("Food Required: %.0f\n", requiredFood)
	fmt.Printf("Free Food: %.0f\n", freeFood)
	fmt.Printf("Additional Food Needed: %.0f\n", foodDifference)
	fmt.Println()
	fmt.Println("1 Gold = 1 Food")
	fmt.Println("Enter the amount of additional Food to purchase.")
	fmt.Println("Leave blank to purchase none.")

	for {
		input, err := readInput(reader, "Food to purchase: ")
		if err != nil {
			fmt.Printf("Could not read Food amount: %v\n", err)
			return 0
		}

		input = strings.TrimSpace(input)

		if input == "" {
			return 0
		}

		amount, err := strconv.Atoi(input)
		if err != nil || amount < 0 {
			fmt.Println("Please enter a valid whole number or leave the entry blank.")
			continue
		}

		return amount
	}
}

func readStartingPosition(
	reader *bufio.Reader,
) ([2]float64, bool) {
	fmt.Println()
	fmt.Println("Starting Locations:")
	fmt.Println("A: (0, 0)")
	fmt.Println("B: (100, 0)")
	fmt.Println("C: (0, 100)")
	fmt.Println("D: (100, 100)")
	fmt.Println("0: Cancel")

	choice, err := readInput(
		reader,
		"Select starting location [A/B/C/D]: ",
	)
	if err != nil {
		fmt.Printf("Could not read starting location: %v\n", err)
		return [2]float64{}, false
	}

	switch strings.ToUpper(strings.TrimSpace(choice)) {
	case "A":
		return [2]float64{0, 0}, true
	case "B":
		return [2]float64{100, 0}, true
	case "C":
		return [2]float64{0, 100}, true
	case "D":
		return [2]float64{100, 100}, true
	case "0":
		return [2]float64{}, false
	default:
		fmt.Println("Invalid starting location.")
		return [2]float64{}, false
	}
}

func readDeploymentDestinations(
	reader *bufio.Reader,
) [][2]float64 {
	const maxDestinations = 5

	destinations := make([][2]float64, 0, maxDestinations)

	for len(destinations) < maxDestinations {
		fmt.Println()

		xInput, err := readInput(
			reader,
			fmt.Sprintf(
				"Destination %d X (0-100, or end to finish): ",
				len(destinations)+1,
			),
		)
		if err != nil {
			fmt.Printf("Could not read X coordinate: %v\n", err)
			return destinations
		}

		xInput = strings.TrimSpace(xInput)
		if strings.EqualFold(xInput, "end") {
			break
		}

		x, err := strconv.ParseFloat(xInput, 64)
		if err != nil || !isWithinWorld(x) {
			fmt.Println("X must be a number from 0 to 100, or end.")
			continue
		}

		yInput, err := readInput(reader, "Destination Y (0-100): ")
		if err != nil {
			fmt.Printf("Could not read Y coordinate: %v\n", err)
			return destinations
		}

		y, err := strconv.ParseFloat(strings.TrimSpace(yInput), 64)
		if err != nil || !isWithinWorld(y) {
			fmt.Println("Y must be a number from 0 to 100.")
			continue
		}

		destinations = append(destinations, [2]float64{x, y})
		fmt.Printf("Added destination (%.2f, %.2f).\n", x, y)

		if len(destinations) < maxDestinations {
			fmt.Println("Enter another destination, or type end when asked for X.")
		}
	}

	if len(destinations) == maxDestinations {
		fmt.Println("Maximum of 5 destinations reached.")
	}

	return destinations
}

func calculateRouteDistance(
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

	// The current movement system retraces the route back home.
	return total * 2.0
}

func readAdditionalFood(
	reader *bufio.Reader,
	currentGold int,
) (int, bool) {
	input, err := readInput(
		reader,
		fmt.Sprintf(
			"Additional Food to purchase [0-%d, blank for 0]: ",
			currentGold,
		),
	)
	if err != nil {
		fmt.Printf("Could not read Food amount: %v\n", err)
		return 0, false
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return 0, true
	}

	amount, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Food must be a whole number.")
		return 0, false
	}

	if amount < 0 || amount > currentGold {
		fmt.Printf(
			"Additional Food must be between 0 and %d.\n",
			currentGold,
		)
		return 0, false
	}

	return amount, true
}

func isWithinWorld(value float64) bool {
	return value >= 0 && value <= 100
}

func isYes(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "y" || value == "yes"
}
