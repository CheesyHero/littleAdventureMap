-- =========================================================
-- BASIC ABILITIES
-- =========================================================

INSERT INTO abilities (
    ability_name,
    ability_type,
    mana_cost,
    accuracy,
    target_type,
    max_targets,
    effect
)
VALUES

    (
        'Spinning Strike',
        'Physical',
        25.0,
        80.0,
        'Enemy',
        3,
        'Damage_30'
    ),

    (
        'Backstab',
        'Physical',
        33.0,
        60.0,
        'Enemy',
        1,
        'Damage_50'
    ),

    (
        'Aim',
        'Physical',
        30.0,
        -1.0,
        'Enemy',
        1,
        'Damage_35'
    ),

    (
        'Cure',
        'Magical',
        25.0,
        -1.0,
        'Friendly',
        1,
        'Heal_25'
    ),

    (
        'First Aid',
        'Physical',
        20.0,
        -1.0,
        'Self',
        1,
        'Heal_20,RemoveDebuff_All'
    ),

    (
        'Guided Bolt',
        'Magical',
        35.0,
        90.0,
        'Enemy',
        1,
        'Damage_30'
    ),

    (
        'Firebolt',
        'Magical',
        20.0,
        90.0,
        'Enemy',
        1,
        'Damage_30'
    ),

    (
        'Lightningbolt',
        'Magical',
        50.0,
        75.0,
        'Enemy',
        3,
        'Damage_50'
    ),

    (
        'Shield',
        'Magical',
        30.0,
        -1.0,
        'Friendly',
        1,
        'Shield_25'
    ),

    (
        'Low Potion',
        'Magical',
        0.0,
        -1.0,
        'Friendly',
        1,
        'Heal_20'
    ),

    (
        'Med Potion',
        'Magical',
        0.0,
        -1.0,
        'Friendly',
        1,
        'Heal_35'
    ),

    (
        'High Potion',
        'Magical',
        0.0,
        -1.0,
        'Friendly',
        1,
        'Heal_60'
    ),

    (
        'Low Ether',
        'Magical',
        0.0,
        -1.0,
        'Friendly',
        1,
        'Mana_20'
    ),

    (
        'Med Ether',
        'Magical',
        0.0,
        -1.0,
        'Friendly',
        1,
        'Mana_35'
    ),

    (
        'Antidote',
        'Magical',
        0.0,
        -1.0,
        'Friendly',
        1,
        'RemoveDebuff_Poison'
    ),

    (
        'Remedy',
        'Magical',
        0.0,
        -1.0,
        'Friendly',
        1,
        'RemoveDebuff_All'
    ),

    (
        'Bomb',
        'Physical',
        0.0,
        85.0,
        'Enemy',
        3,
        'Damage_25'
    ),

    (
        'Poison Flask',
        'Physical',
        0.0,
        85.0,
        'Enemy',
        1,
        'Poison_3'
    )

ON CONFLICT (ability_name) DO NOTHING;