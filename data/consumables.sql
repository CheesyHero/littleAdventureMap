-- =========================================================
-- STARTER CONSUMABLES
-- =========================================================

INSERT INTO consumables (
    consumable_name,
    gold_cost,
    ability_name
)
VALUES

    (
        'Low Potion',
        25,
        'Low Potion'
    ),

    (
        'Med Potion',
        50,
        'Med Potion'
    ),

    (
        'High Potion',
        100,
        'High Potion'
    ),

    (
        'Low Ether',
        30,
        'Low Ether'
    ),

    (
        'Med Ether',
        60,
        'Med Ether'
    ),

    (
        'Antidote',
        20,
        'Antidote'
    ),

    (
        'Remedy',
        75,
        'Remedy'
    ),

    (
        'Bomb',
        40,
        'Bomb'
    ),

    (
        'Poison Flask',
        45,
        'Poison Flask'
    )

ON CONFLICT (consumable_name) DO NOTHING;