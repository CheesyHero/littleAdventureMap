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

- Usage

Little Adventure Map has two command interfaces: the client, used to play the game, and the server console, used to monitor and manage the simulation.

Client Commands

1:	Hire a Hero

2:	Deploy a Hero on a journey

3:	Manage Heroes and view adventure logs

4:	View Inventory

5:	Open the Market

6:	Open the Library

7:	View the live world map

8:	Quit

The live map displays each deployed hero as a uniquely colored marker and refreshes every second as the server advances the simulation. Press any key to return to the main menu.

Server Commands

The server also provides an admin console for inspecting and controlling the simulation.

0: help	Show available commands

1: add-user	Create a test user

2: print-users	View all users

e: edit-user	Edit user progression

3: add-agent	Create an agent

4: print-agents	View all agents

a: edit-agent	Edit agent stats

5: set-destination	Deploy an agent to a destination

6: print-destinations	View an agent's route and status

7: watch-agents	Watch agent movement live

p: pause	Pause or resume the simulation

8: withdraw-agent	Withdraw an agent from its journey

9: exit	Shut down the server

The server's watch-agents command provides a live debugging view of the simulation, while the client's map provides the player-facing view of their own heroes.
