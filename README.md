How it all works!
The project requires a postgres database.
There are two main.go files:

server/main.go
client/main.go

Launching the server first will initialize the psql database with all the proper data required.
When the server is running, it ticks every one second.

Launching a client will ask for a username and password.
If the entered one does not exist, it will ask if a new account should be made.

As a new user, you need to hire a hero with the "1" command. The first one is free.
You can then start a journey with the "2" command.
When prompted, select a hero that is available and at home.
Pick a starting corner from the map to start.
Set coordinates to go to.
You will be informed of the projected food cost and how much food to buy for your hero.
Once committed, the hero will journey on their own for a few minutes depending on the length of their journey.

You can evaluate their journey as they go and check up on the logs they print in the "manage heroes" menu.

And that's the game!
Sometimes your heroes find food and a bunch of extra gold.
Sometimes they get their butts kicked by a bunch of goblins. 

The core elements are all in place.
I actually plan to update the following elements:

Agents can interact with each other and leave logs.
A combat system that can incorporate multiple participants and alliances.
Items such as equipment and consumables you can give your heroes, like antidotes to fight spiders.
