-- =========================================================
-- USERS
-- =========================================================

CREATE TABLE users (
    user_id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    passwordhash TEXT NOT NULL,

    money INT NOT NULL DEFAULT 1000
        CHECK (money >= 0),

    max_agents INT NOT NULL DEFAULT 1
        CHECK (max_agents >= 0),

    guild_level INT NOT NULL DEFAULT 1
        CHECK (guild_level >= 1),

    guild_experience DOUBLE PRECISION NOT NULL DEFAULT 0.0
        CHECK (guild_experience >= 0.0),

    guild_next_level_exp INT
        GENERATED ALWAYS AS (
            (guild_level + 1) * (guild_level + 1) * (guild_level + 1)
        ) STORED,

    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- =========================================================
-- ABILITIES
-- =========================================================

CREATE TABLE abilities (
    ability_name TEXT PRIMARY KEY,

    ability_type TEXT NOT NULL,

    mana_cost DOUBLE PRECISION NOT NULL DEFAULT 0.0
        CHECK (mana_cost >= 0.0),

    -- -1 = guaranteed hit
    accuracy DOUBLE PRECISION NOT NULL DEFAULT 100.0
        CHECK (
            accuracy = -1.0
            OR (accuracy >= 0.0 AND accuracy <= 100.0)
        ),

    target_type TEXT NOT NULL CHECK (
        target_type IN (
            'Self',
            'Enemy',
            'Friendly',
            'Other',
            'Any'
        )
    ),

    max_targets INT NOT NULL DEFAULT 1
        CHECK (max_targets >= 1),

    -- Parsed later by game logic.
    -- Examples: Damage_30, Heal_25, Poison_3
    effect TEXT NOT NULL
);


-- =========================================================
-- EQUIPMENT
-- =========================================================

CREATE TABLE equipment (
    equipment_name TEXT PRIMARY KEY,

    equipment_type TEXT NOT NULL CHECK (
        equipment_type IN (
            'Weapon',
            'Armor',
            'Accessory'
        )
    ),

    gold_cost INT NOT NULL DEFAULT 100
        CHECK (gold_cost >= 0),

    max_health DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    health_regen DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    max_mana DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    mana_regen DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    food_cost DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    speed DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    strength DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    endurance DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    intelligence DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    wisdom DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    agility DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    charisma DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    defense DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    resist DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    critical_chance DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    critical_damage DOUBLE PRECISION NOT NULL DEFAULT 0.0
);


-- =========================================================
-- CONSUMABLES
-- =========================================================

CREATE TABLE consumables (
    consumable_name TEXT PRIMARY KEY,

    gold_cost INT NOT NULL DEFAULT 0
        CHECK (gold_cost >= 0),

    ability_name TEXT NOT NULL
        REFERENCES abilities(ability_name)
);


-- =========================================================
-- CLASSES
-- =========================================================

CREATE TABLE classes (
    class_name TEXT PRIMARY KEY,

    -- Level 1 Stats
    base_max_health DOUBLE PRECISION NOT NULL,
    base_health_regen DOUBLE PRECISION NOT NULL,

    base_max_mana DOUBLE PRECISION NOT NULL,
    base_mana_regen DOUBLE PRECISION NOT NULL,

    base_food_cost DOUBLE PRECISION NOT NULL,
    base_speed DOUBLE PRECISION NOT NULL,

    base_strength DOUBLE PRECISION NOT NULL,
    base_endurance DOUBLE PRECISION NOT NULL,
    base_intelligence DOUBLE PRECISION NOT NULL,
    base_wisdom DOUBLE PRECISION NOT NULL,
    base_agility DOUBLE PRECISION NOT NULL,
    base_charisma DOUBLE PRECISION NOT NULL,

    base_defense DOUBLE PRECISION NOT NULL,
    base_resist DOUBLE PRECISION NOT NULL,

    base_critical_chance DOUBLE PRECISION NOT NULL,
    base_critical_damage DOUBLE PRECISION NOT NULL,

    -- Growth Per Level Above 1
    per_level_max_health DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    per_level_health_regen DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    per_level_max_mana DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    per_level_mana_regen DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    per_level_food_cost DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    per_level_speed DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    per_level_strength DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    per_level_endurance DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    per_level_intelligence DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    per_level_wisdom DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    per_level_agility DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    per_level_charisma DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    per_level_defense DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    per_level_resist DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    per_level_critical_chance DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    per_level_critical_damage DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    -- Starting Equipment
    base_weapon TEXT
        REFERENCES equipment(equipment_name),

    base_armor_equipment TEXT
        REFERENCES equipment(equipment_name)
);


-- =========================================================
-- CLASS ABILITIES
-- =========================================================

CREATE TABLE class_abilities (
    class_name TEXT NOT NULL
        REFERENCES classes(class_name)
        ON DELETE CASCADE,

    ability_name TEXT NOT NULL
        REFERENCES abilities(ability_name)
        ON DELETE CASCADE,

    PRIMARY KEY (class_name, ability_name)
);


-- =========================================================
-- AGENTS / HEROES
-- =========================================================

CREATE TABLE agents (
    agent_id BIGSERIAL PRIMARY KEY,

    owner_id BIGINT NOT NULL
        REFERENCES users(user_id)
        ON DELETE CASCADE,

    agentname TEXT NOT NULL,

    class_name TEXT
        REFERENCES classes(class_name),

    -- Deployment
    deployed BOOLEAN NOT NULL DEFAULT FALSE,

    -- Gold collected during the current journey.
    journey_gold INT NOT NULL DEFAULT 0
        CHECK (journey_gold >= 0),

    -- NULL while the hero is not on the map.
    world_position DOUBLE PRECISION[],
    home_position DOUBLE PRECISION[],

    path_index INT NOT NULL DEFAULT 0
        CHECK (path_index >= 0),

    returning_home BOOLEAN NOT NULL DEFAULT FALSE,

    -- Travel
    speed DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    food_cost DOUBLE PRECISION NOT NULL DEFAULT 1.0,

    food_owned DOUBLE PRECISION NOT NULL DEFAULT 100.0
        CHECK (food_owned >= -20.0),

    -- Health
    current_health DOUBLE PRECISION NOT NULL DEFAULT 100.0,
    max_health DOUBLE PRECISION NOT NULL DEFAULT 100.0,
    health_regen DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    -- Mana
    current_mana DOUBLE PRECISION NOT NULL DEFAULT 100.0,
    max_mana DOUBLE PRECISION NOT NULL DEFAULT 100.0,
    mana_regen DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    -- Primary Attributes
    strength DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    endurance DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    intelligence DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    wisdom DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    agility DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    charisma DOUBLE PRECISION NOT NULL DEFAULT 1.0,

    -- Secondary Attributes
    defense DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    resist DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    critical_chance DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    critical_damage DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    -- Progression
    experience DOUBLE PRECISION NOT NULL DEFAULT 0.0
        CHECK (experience >= 0.0),

    level INT NOT NULL DEFAULT 1
        CHECK (level >= 1),

    next_level_exp INT
        GENERATED ALWAYS AS (
            (level + 1) * (level + 1) * (level + 1)
        ) STORED
);


-- =========================================================
-- EQUIPPED ITEMS
--
-- Maximum of four equipped items:
-- Weapon, Armor, Accessory1, Accessory2
-- =========================================================

CREATE TABLE agent_equipment (
    agent_id BIGINT NOT NULL
        REFERENCES agents(agent_id)
        ON DELETE CASCADE,

    equipment_name TEXT NOT NULL
        REFERENCES equipment(equipment_name),

    equipment_slot TEXT NOT NULL CHECK (
        equipment_slot IN (
            'Weapon',
            'Armor',
            'Accessory1',
            'Accessory2'
        )
    ),

    PRIMARY KEY (agent_id, equipment_slot)
);


-- =========================================================
-- USER INVENTORY
-- =========================================================

CREATE TABLE user_equipment_inventory (
    user_id BIGINT NOT NULL
        REFERENCES users(user_id)
        ON DELETE CASCADE,

    equipment_name TEXT NOT NULL
        REFERENCES equipment(equipment_name),

    quantity INT NOT NULL DEFAULT 1
        CHECK (quantity > 0),

    PRIMARY KEY (user_id, equipment_name)
);


CREATE TABLE user_consumable_inventory (
    user_id BIGINT NOT NULL
        REFERENCES users(user_id)
        ON DELETE CASCADE,

    consumable_name TEXT NOT NULL
        REFERENCES consumables(consumable_name),

    quantity INT NOT NULL DEFAULT 1
        CHECK (quantity > 0),

    PRIMARY KEY (user_id, consumable_name)
);


-- =========================================================
-- AGENT INVENTORY
--
-- Capacity rules are handled by game logic.
-- =========================================================

CREATE TABLE agent_equipment_inventory (
    agent_id BIGINT NOT NULL
        REFERENCES agents(agent_id)
        ON DELETE CASCADE,

    equipment_name TEXT NOT NULL
        REFERENCES equipment(equipment_name),

    quantity INT NOT NULL DEFAULT 1
        CHECK (quantity > 0),

    PRIMARY KEY (agent_id, equipment_name)
);


CREATE TABLE agent_consumable_inventory (
    agent_id BIGINT NOT NULL
        REFERENCES agents(agent_id)
        ON DELETE CASCADE,

    consumable_name TEXT NOT NULL
        REFERENCES consumables(consumable_name),

    quantity INT NOT NULL DEFAULT 1
        CHECK (quantity > 0),

    PRIMARY KEY (agent_id, consumable_name)
);


-- =========================================================
-- DESTINATIONS
-- =========================================================

CREATE TABLE destinations (
    destination_id BIGSERIAL PRIMARY KEY,

    agent_id BIGINT NOT NULL
        REFERENCES agents(agent_id)
        ON DELETE CASCADE,

    destination_order INT NOT NULL DEFAULT 0
        CHECK (destination_order >= 0),

    destination_name TEXT NOT NULL,

    world_position DOUBLE PRECISION[] NOT NULL,

    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (agent_id, destination_order)
);


-- =========================================================
-- ADVENTURE LOGS
-- =========================================================

CREATE TABLE event_results (
    event_id BIGSERIAL PRIMARY KEY,

    agent_id BIGINT NOT NULL
        REFERENCES agents(agent_id)
        ON DELETE CASCADE,

    title TEXT NOT NULL,
    description TEXT NOT NULL,
    results TEXT NOT NULL,

    where_position DOUBLE PRECISION[] NOT NULL,

    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- =========================================================
-- MONSTERS
--
-- Runtime combat entities.
-- =========================================================

CREATE TABLE monsters (
    monster_id BIGSERIAL PRIMARY KEY,

    monster_name TEXT NOT NULL,

    class_name TEXT NOT NULL
        REFERENCES classes(class_name),

    world_position DOUBLE PRECISION[],

    speed DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    food_cost DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    -- Health
    current_health DOUBLE PRECISION NOT NULL,
    max_health DOUBLE PRECISION NOT NULL,
    health_regen DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    -- Mana
    current_mana DOUBLE PRECISION NOT NULL,
    max_mana DOUBLE PRECISION NOT NULL,
    mana_regen DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    -- Primary Attributes
    strength DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    endurance DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    intelligence DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    wisdom DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    agility DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    charisma DOUBLE PRECISION NOT NULL DEFAULT 1.0,

    -- Secondary Attributes
    defense DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    resist DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    critical_chance DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    critical_damage DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    -- Progression
    experience DOUBLE PRECISION NOT NULL DEFAULT 0.0
        CHECK (experience >= 0.0),

    level INT NOT NULL DEFAULT 1
        CHECK (level >= 1),

    next_level_exp INT
        GENERATED ALWAYS AS (
            (level + 1) * (level + 1) * (level + 1)
        ) STORED,

    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- =========================================================
-- PARTIES
-- =========================================================

CREATE TABLE parties (
    party_id BIGSERIAL PRIMARY KEY,

    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE party_agents (
    party_id BIGINT NOT NULL
        REFERENCES parties(party_id)
        ON DELETE CASCADE,

    agent_id BIGINT NOT NULL
        REFERENCES agents(agent_id)
        ON DELETE CASCADE,

    PRIMARY KEY (party_id, agent_id),

    -- An Agent may only belong to one active party.
    UNIQUE (agent_id)
);


CREATE TABLE party_monsters (
    party_id BIGINT NOT NULL
        REFERENCES parties(party_id)
        ON DELETE CASCADE,

    monster_id BIGINT NOT NULL
        REFERENCES monsters(monster_id)
        ON DELETE CASCADE,

    PRIMARY KEY (party_id, monster_id),

    -- A Monster may only belong to one active party.
    UNIQUE (monster_id)
);


-- =========================================================
-- ENGAGEMENTS
-- =========================================================

CREATE TABLE engagements (
    engagement_id BIGSERIAL PRIMARY KEY,

    world_position DOUBLE PRECISION[] NOT NULL,

    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE engagement_parties (
    engagement_id BIGINT NOT NULL
        REFERENCES engagements(engagement_id)
        ON DELETE CASCADE,

    party_id BIGINT NOT NULL
        REFERENCES parties(party_id)
        ON DELETE CASCADE,

    PRIMARY KEY (engagement_id, party_id),

    -- A Party may only participate in one active engagement.
    UNIQUE (party_id)
);


-- =========================================================
-- INDEXES
-- =========================================================

CREATE INDEX idx_agents_owner
    ON agents(owner_id);

CREATE INDEX idx_agents_deployed
    ON agents(deployed);

CREATE INDEX idx_agents_class
    ON agents(class_name);

CREATE INDEX idx_destinations_agent
    ON destinations(agent_id);

CREATE INDEX idx_event_results_agent
    ON event_results(agent_id);

CREATE INDEX idx_event_results_timestamp
    ON event_results(timestamp);

CREATE INDEX idx_monsters_class
    ON monsters(class_name);

CREATE INDEX idx_party_agents_party
    ON party_agents(party_id);

CREATE INDEX idx_party_monsters_party
    ON party_monsters(party_id);

CREATE INDEX idx_engagement_parties_engagement
    ON engagement_parties(engagement_id);