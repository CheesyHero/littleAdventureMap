# Little Adventure Map

An asynchronous hero-simulation game built with Go and PostgreSQL. Players recruit heroes, send them on journeys, and check back later to see what happened. The server advances the world state every second, resolving active journeys and recording persistent event logs to PostgreSQL.

- Motivation

I wanted to build a game where progress happens over time rather than immediately after a button press. That led me to the idea of sending heroes on journeys that continue independently while a server simulates their progress and records what happens. Little Adventure Map became a way for me to explore persistent state, asynchronous gameplay, and database-driven systems while still building a fun and enjoyable experience.

- Quick Start

Make sure PostgreSQL is installed and running, then start the server:

go run server/main.go

The server will initialize the game database and begin running the simulation.

In a second terminal, start the client:

go run client/main.go

Log in or create a new account, hire your first hero, and start a journey.

- Technical Overview

The project uses a separate client and server so the simulation can continue independently of player input. PostgreSQL stores persistent users, heroes, journey state, and event logs.

The main engineering challenges came from keeping the simulation and database state synchronized while enforcing gameplay restrictions. Individual systems were fairly straightforward to build, but handling database constraints, state transitions, and repeated updates from the simulation required considerably more iteration.

- Running the Project

Requires Go and PostgreSQL.

Set your DATABASE_URL and start the server:

go run server/main.go

In another terminal, start the client:

go run client/main.go
Create an account, hire a hero, and send them on a journey.
Use command 7 to open the live map and watch your deployed heroes move across the world.

- Roadmap

Future work includes hero-to-hero interactions, multi-participant combat and alliances, equipment, and consumable items such as antidotes that can help heroes prepare for specific encounters.

The core gameplay loop is already in place.
