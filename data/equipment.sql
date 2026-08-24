-- =========================================================
-- STARTER WEAPONS
-- =========================================================

INSERT INTO equipment (
    equipment_name,
    equipment_type,
    strength
)
VALUES
    ('Iron Pick', 'Weapon', 2.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    agility
)
VALUES
    ('Short Bow', 'Weapon', 2.0),
    ('Shiv', 'Weapon', 1.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    intelligence
)
VALUES
    ('Quartz Wand', 'Weapon', 2.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    wisdom
)
VALUES
    ('Oak Staff', 'Weapon', 2.0)
ON CONFLICT (equipment_name) DO NOTHING;


-- =========================================================
-- STARTER ARMOR
-- =========================================================

INSERT INTO equipment (
    equipment_name,
    equipment_type,
    defense
)
VALUES
    ('Cloth Robes', 'Armor', 1.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    defense,
    agility
)
VALUES
    ('Padded Armor', 'Armor', 1.0, 1.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    defense,
    agility
)
VALUES
    ('Chain Mail', 'Armor', 3.0, -1.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    defense,
    agility
)
VALUES
    ('Wooden Plate Mail', 'Armor', 5.0, -2.0)
ON CONFLICT (equipment_name) DO NOTHING;


-- =========================================================
-- STARTER ACCESSORIES / HEADGEAR
-- =========================================================

INSERT INTO equipment (
    equipment_name,
    equipment_type,
    strength
)
VALUES
    ('Headband', 'Accessory', 1.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    charisma
)
VALUES
    ('Barret', 'Accessory', 1.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    intelligence
)
VALUES
    ('Circlet', 'Accessory', 1.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    wisdom
)
VALUES
    ('Cloth Hood', 'Accessory', 1.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    agility
)
VALUES
    ('Leather Cap', 'Accessory', 1.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    defense
)
VALUES
    ('Chain Hood', 'Accessory', 2.0)
ON CONFLICT (equipment_name) DO NOTHING;


INSERT INTO equipment (
    equipment_name,
    equipment_type,
    defense,
    agility
)
VALUES
    ('Wooden Full Helm', 'Accessory', 4.0, -1.0)
ON CONFLICT (equipment_name) DO NOTHING;