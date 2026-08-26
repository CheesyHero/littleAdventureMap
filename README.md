# Little Adventure Map

An asynchronous hero-simulation game built with Go and PostgreSQL. Players recruit heroes, send them on journeys, and check back later to see what happened. The server advances the world state every second, resolving active journeys and recording persistent event logs to PostgreSQL.

- Motivation

I wanted to build a game where progress happens over time rather than immediately after a button press. That led me to the idea of sending heroes on journeys that continue independently while a server simulates their progress and records what happens. Little Adventure Map became a way for me to explore persistent state, asynchronous gameplay, and database-driven systems while still building a fun an enjoyable experience.

- How It Works

The project has two main programs:

server/main.go
client/main.go

The server initializes the PostgreSQL database and runs the simulation. While it is running, deployed heroes continue traveling and resolving events every second.

go run server/main.go

The client handles accounts, hero management, journeys, and adventure logs.

go run client/main.go

When launching the client, enter a username and password. If the account does not exist, you can create one.

- Gameplay

New players can hire their first hero for free using the 1 command.

Use 2 to start a journey. Select an available hero, choose a starting corner, and enter a destination. The game will estimate the food required before you commit to the journey.

Once deployed, the hero travels independently for several minutes depending on the route.

You can check their progress and read the events generated during the journey through the Manage Heroes menu.

- Technical Overview

The project uses a separate client and server so the simulation can continue independently of player input. PostgreSQL stores persistent users, heroes, journey state, and event logs.
Some of the main engineering challenges have making sure all the syntax is correct. The logic and game data was easy to set up and get working, but saving and iterating to the database when something had a limit or restriction caused some headaches.

- Running the Project

Requires Go and PostgreSQL.

1: go run server/main.go

2: go run client/main.go

3: Create an account, hire a hero, and send them on a journey.

- Roadmap

Future work includes hero-to-hero interactions, multi-participant combat and alliances, equipment, and consumable items such as antidotes that can help heroes prepare for specific encounters.

The core gameplay loop is already in place.
