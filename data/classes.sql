-- =========================================================
-- WARRIOR
-- =========================================================

INSERT INTO classes (
    class_name,

    base_max_health, base_health_regen,
    base_max_mana, base_mana_regen,
    base_food_cost, base_speed,

    base_strength, base_endurance, base_intelligence,
    base_wisdom, base_agility, base_charisma,

    base_defense, base_resist,
    base_critical_chance, base_critical_damage,

    per_level_max_health, per_level_health_regen,
    per_level_max_mana, per_level_mana_regen,
    per_level_food_cost, per_level_speed,

    per_level_strength, per_level_endurance, per_level_intelligence,
    per_level_wisdom, per_level_agility, per_level_charisma,

    per_level_defense, per_level_resist,
    per_level_critical_chance, per_level_critical_damage,

    base_weapon, base_armor_equipment
)
VALUES (
    'Warrior',

    140.0, 0.25,
    80.0, 0.25,
    1.1, 1.1,

    8.0, 6.0, 4.0,
    4.0, 5.0, 5.0,

    1.0, 1.0,
    0.05, 1.50,

    8.0, 0.05,
    2.0, 0.025,
    0.0, 0.0,

    1.0, 1.0, 0.20,
    0.25, 0.40, 0.25,

    0.0, 0.0,
    0.0, 0.0,

    'Iron Pick',
    'Chain Mail'
)
ON CONFLICT (class_name) DO NOTHING;


-- =========================================================
-- ROGUE
-- =========================================================

INSERT INTO classes (
    class_name,

    base_max_health, base_health_regen,
    base_max_mana, base_mana_regen,
    base_food_cost, base_speed,

    base_strength, base_endurance, base_intelligence,
    base_wisdom, base_agility, base_charisma,

    base_defense, base_resist,
    base_critical_chance, base_critical_damage,

    per_level_max_health, per_level_health_regen,
    per_level_max_mana, per_level_mana_regen,
    per_level_food_cost, per_level_speed,

    per_level_strength, per_level_endurance, per_level_intelligence,
    per_level_wisdom, per_level_agility, per_level_charisma,

    per_level_defense, per_level_resist,
    per_level_critical_chance, per_level_critical_damage,

    base_weapon, base_armor_equipment
)
VALUES (
    'Rogue',

    90.0, 0.20,
    80.0, 0.33,
    0.95, 1.20,

    5.0, 4.0, 5.0,
    4.0, 8.0, 6.0,

    1.0, 1.0,
    0.06, 1.75,

    4.0, 0.033,
    5.5, 0.0425,
    0.0, 0.0,

    0.45, 0.35, 0.30,
    0.20, 1.10, 0.50,

    0.0, 0.0,
    0.0, 0.0,

    'Shiv',
    'Padded Armor'
)
ON CONFLICT (class_name) DO NOTHING;


-- =========================================================
-- RANGER
-- =========================================================

INSERT INTO classes (
    class_name,

    base_max_health, base_health_regen,
    base_max_mana, base_mana_regen,
    base_food_cost, base_speed,

    base_strength, base_endurance, base_intelligence,
    base_wisdom, base_agility, base_charisma,

    base_defense, base_resist,
    base_critical_chance, base_critical_damage,

    per_level_max_health, per_level_health_regen,
    per_level_max_mana, per_level_mana_regen,
    per_level_food_cost, per_level_speed,

    per_level_strength, per_level_endurance, per_level_intelligence,
    per_level_wisdom, per_level_agility, per_level_charisma,

    per_level_defense, per_level_resist,
    per_level_critical_chance, per_level_critical_damage,

    base_weapon, base_armor_equipment
)
VALUES (
    'Ranger',

    100.0, 0.20,
    80.0, 0.40,
    0.80, 1.15,

    7.0, 4.0, 5.0,
    5.0, 7.0, 4.0,

    1.0, 1.0,
    0.06, 1.75,

    5.0, 0.04,
    5.0, 0.04,
    0.0, 0.0,

    1.0, 0.45, 0.30,
    0.40, 1.0, 0.50,

    0.0, 0.0,
    0.0, 0.0,

    'Short Bow',
    'Padded Armor'
)
ON CONFLICT (class_name) DO NOTHING;


-- =========================================================
-- WIZARD
-- =========================================================

INSERT INTO classes (
    class_name,

    base_max_health, base_health_regen,
    base_max_mana, base_mana_regen,
    base_food_cost, base_speed,

    base_strength, base_endurance, base_intelligence,
    base_wisdom, base_agility, base_charisma,

    base_defense, base_resist,
    base_critical_chance, base_critical_damage,

    per_level_max_health, per_level_health_regen,
    per_level_max_mana, per_level_mana_regen,
    per_level_food_cost, per_level_speed,

    per_level_strength, per_level_endurance, per_level_intelligence,
    per_level_wisdom, per_level_agility, per_level_charisma,

    per_level_defense, per_level_resist,
    per_level_critical_chance, per_level_critical_damage,

    base_weapon, base_armor_equipment
)
VALUES (
    'Wizard',

    80.0, 0.125,
    100.0, 0.50,
    1.0, 1.0,

    4.0, 4.0, 8.0,
    6.0, 5.0, 5.0,

    1.0, 1.0,
    0.05, 1.50,

    3.0, 0.02,
    8.0, 0.25,
    0.0, 0.0,

    0.15, 0.20, 1.20,
    0.70, 0.25, 0.25,

    0.0, 0.0,
    0.0, 0.0,

    'Quartz Wand',
    'Cloth Robes'
)
ON CONFLICT (class_name) DO NOTHING;


-- =========================================================
-- CLERIC
-- =========================================================

INSERT INTO classes (
    class_name,

    base_max_health, base_health_regen,
    base_max_mana, base_mana_regen,
    base_food_cost, base_speed,

    base_strength, base_endurance, base_intelligence,
    base_wisdom, base_agility, base_charisma,

    base_defense, base_resist,
    base_critical_chance, base_critical_damage,

    per_level_max_health, per_level_health_regen,
    per_level_max_mana, per_level_mana_regen,
    per_level_food_cost, per_level_speed,

    per_level_strength, per_level_endurance, per_level_intelligence,
    per_level_wisdom, per_level_agility, per_level_charisma,

    per_level_defense, per_level_resist,
    per_level_critical_chance, per_level_critical_damage,

    base_weapon, base_armor_equipment
)
VALUES (
    'Cleric',

    100.0, 1.0,
    100.0, 0.50,
    0.90, 1.0,

    4.0, 6.0, 5.0,
    8.0, 4.0, 5.0,

    1.0, 1.0,
    0.05, 1.50,

    4.0, 0.10,
    6.0, 0.10,
    0.0, 0.0,

    0.30, 0.60, 0.40,
    1.0, 0.20, 0.50,

    0.0, 0.0,
    0.0, 0.0,

    'Oak Staff',
    'Cloth Robes'
)
ON CONFLICT (class_name) DO NOTHING;


-- =========================================================
-- GOBLIN
-- =========================================================

INSERT INTO classes (
    class_name,

    base_max_health, base_health_regen,
    base_max_mana, base_mana_regen,
    base_food_cost, base_speed,

    base_strength, base_endurance, base_intelligence,
    base_wisdom, base_agility, base_charisma,

    base_defense, base_resist,
    base_critical_chance, base_critical_damage,

    per_level_max_health, per_level_health_regen,
    per_level_max_mana, per_level_mana_regen,
    per_level_food_cost, per_level_speed,

    per_level_strength, per_level_endurance, per_level_intelligence,
    per_level_wisdom, per_level_agility, per_level_charisma,

    per_level_defense, per_level_resist,
    per_level_critical_chance, per_level_critical_damage,

    base_weapon, base_armor_equipment
)
VALUES (
    'Goblin',

    50.0, 0.10,
    50.0, 0.25,
    1.0, 1.0,

    3.0, 2.0, 1.0,
    1.0, 3.0, 1.0,

    1.0, 1.0,
    0.05, 1.50,

    5.0, 0.05,
    5.0, 0.05,
    0.0, 0.0,

    1.0, 0.75, 0.25,
    0.30, 1.0, 0.25,

    0.10, 0.10,
    0.0, 0.0,

    'Shiv',
    'Cloth Robes'
)
ON CONFLICT (class_name) DO NOTHING;


-- =========================================================
-- STARTING CLASS ABILITIES
--
-- abilities.sql must be loaded before this file.
-- Goblin intentionally has no starting abilities.
-- =========================================================

INSERT INTO class_abilities (
    class_name,
    ability_name
)
VALUES
    -- Warrior
    ('Warrior', 'Spinning Strike'),
    ('Warrior', 'First Aid'),

    -- Rogue
    ('Rogue', 'Backstab'),

    -- Ranger
    ('Ranger', 'Aim'),
    ('Ranger', 'First Aid'),

    -- Wizard
    ('Wizard', 'Firebolt'),
    ('Wizard', 'Lightningbolt'),
    ('Wizard', 'Shield'),

    -- Cleric
    ('Cleric', 'Cure'),
    ('Cleric', 'Shield'),
    ('Cleric', 'Guided Bolt')

ON CONFLICT (class_name, ability_name) DO NOTHING;