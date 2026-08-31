# Little Adventure Map

A persistent adventure simulation built with **Go and PostgreSQL**.

Players recruit heroes, send them on journeys, and check back later to see what happened. A continuously running server advances the world every second, moving heroes across the map, consuming resources, generating events, resolving combat, and recording their history.

## Why I Built It

I wanted to experiment with a game where progress happens over time instead of immediately after a button press.

That simple idea turned into a project focused on:

* Persistent simulation
* Database-backed game state
* Resource management
* Procedural events
* Turn-based combat
* Client/server architecture
* Historical event tracking

The game itself is intentionally small. The interesting part is keeping a continuously changing world consistent and observable.

## Features

* Persistent tick-based simulation
* Multiple hero classes with different stats and progression
* Multi-stop journeys across a coordinate-based world
* Food, health, mana, gold, and experience management
* Random events with rewards and penalties
* Weighted monster encounters
* D20-inspired combat
* Persistent combat and adventure logs
* Guild and hero progression
* Live terminal map
* Separate Go client and simulation server
* PostgreSQL-backed state

## Architecture

```text
Player Client
     │
     ▼
 PostgreSQL
     ▲
     │
Simulation Server
     │
     ├── Movement
     ├── Random Events
     ├── Combat
     └── Progression
```

The **client** handles player interaction.

The **server** runs independently and advances the world once per second.

**PostgreSQL** acts as the persistent source of truth shared by both processes.

## Simulation

Each server tick processes the world in phases:

```text
Tick
 │
 ├── Movement
 │     Advance deployed heroes
 │
 ├── Events
 │     Consume resources
 │     Apply regeneration
 │     Roll random encounters
 │
 └── Combat
       Resolve turns
       Remove defeated combatants
       Award rewards
       Record history
```

Separating these phases makes state transitions easier to reason about and prevents newly created events from unexpectedly resolving during the same phase.

## Journeys

Heroes can be assigned several destinations before leaving home.

Longer journeys consume more food and expose the hero to more events, creating a tradeoff between risk and potential rewards.

During an adventure a hero may:

* Find or lose food
* Discover gold
* Take damage
* Recover health or mana
* Encounter monsters
* Gain experience
* Return home with accumulated rewards

## Combat

Monster encounters create persistent engagements between parties.

Combat currently includes:

* Agility-based turn order
* D20-style hit rolls
* Strength and defense-based damage
* Multiple combatants
* Defending and fleeing
* Defeat handling
* Experience and gold rewards

When a combatant is defeated, their rewards are added to the engagement. When combat ends, the total rewards are divided among the surviving heroes.

Combat is also recorded in each participating hero's adventure log.

```text
Combat
Quincy dealt 18.42 damage to Orc.

Combat
Orc was defeated.

Combat
Quincy emerged victorious.
Gained 20 EXP and 50 Gold.
```

## Live Map

The client includes a live terminal map that refreshes while the simulation runs.

```text
World Map

100 |.........................................|
    |......................●..................|
    |.........................................|
    |............!!...........................|
    |.........................................|
  0 |.........................................|
     -----------------------------------------
     0                                     100
```

Each hero has their own marker, while heroes currently in combat are highlighted separately.

This makes the asynchronous simulation visible instead of leaving everything hidden in database state and logs.

## Running Locally

Requires:

* Go
* PostgreSQL

Clone the project:

```bash
git clone <repository-url>
cd littleAdventureMap
```

Install dependencies:

```bash
go mod download
```

Set your database connection:

```bash
export DATABASE_URL="your-postgresql-connection-string"
```

Start the server:

```bash
go run server/main.go
```

Then start the client in another terminal:

```bash
go run client/main.go
```

Create an account, recruit a hero, and send them on an adventure.

Use **View Map** from the client to watch deployed heroes move through the world.

## Development

Build:

```bash
go build ./...
```

Test:

```bash
go test ./...
```

Format:

```bash
gofmt -w .
```

The server also includes an admin console for inspecting users, heroes, destinations, and live simulation state.

## Engineering Lessons

The biggest challenge has not been any individual game mechanic, but coordinating state across a continuously running simulation and a relational database.

Areas that required the most iteration include:

* Keeping simulation and database state synchronized
* Safely transitioning heroes between travel, combat, defeat, and return states
* Managing relationships between parties and engagements
* Handling rewards without losing or duplicating state
* Making asynchronous behavior observable for debugging

## Next Steps

Planned improvements include:

* Deterministic simulations using configurable random seeds
* More transactional combat and reward updates
* Protection against multiple servers processing the same state
* Structured logging and simulation metrics
* More deterministic tests around combat and journey rules
* Expanded items, abilities, and inventory systems

## About the Project

I always wanted to both build and idle game and learn about networking and servers so I could play with my friends, and this was my solution!Building around that idea gave me a practical way to explore persistent state, asynchronous systems, simulation design, PostgreSQL, and Go while still making something interactive and fun. My next goal is to build a GUI for the project and take it online.
