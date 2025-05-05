CREATE TABLE IF NOT EXISTS lemons
(
    position_id       INTEGER PRIMARY KEY CHECK (position_id BETWEEN 0 AND 9),
    is_available      BOOLEAN                  NOT NULL DEFAULT false,
    last_harvested_at TIMESTAMP WITH TIME ZONE,
    next_available_at TIMESTAMP WITH TIME ZONE,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

INSERT INTO lemons (position_id, is_available)
VALUES (0, false),
       (1, false),
       (2, false),
       (3, false),
       (4, false),
       (5, false),
       (6, false),
       (7, false),
       (8, false),
       (9, false)
ON CONFLICT (position_id) DO NOTHING;