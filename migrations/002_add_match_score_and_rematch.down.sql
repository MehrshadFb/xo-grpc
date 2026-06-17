ALTER TABLE games
    DROP COLUMN IF EXISTS rematch_o_requested,
    DROP COLUMN IF EXISTS rematch_x_requested,
    DROP COLUMN IF EXISTS round_number,
    DROP COLUMN IF EXISTS draws,
    DROP COLUMN IF EXISTS o_wins,
    DROP COLUMN IF EXISTS x_wins;
