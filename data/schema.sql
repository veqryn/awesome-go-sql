CREATE TYPE COLORS AS ENUM('red', 'green', 'blue');

CREATE TABLE accounts (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50)              NOT NULL,
    email       VARCHAR(50) UNIQUE       NOT NULL,
    active      BOOLEAN                  NOT NULL,
    fav_color   COLORS,
    fav_numbers INTEGER[],
    properties  JSONB,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL
);

INSERT INTO accounts (id, name, email, active, fav_color, fav_numbers, properties, created_at)
VALUES (1, 'Bob', 'bob@internal.com', true, 'red', '{5}', '{"tags": ["fun"]}', '2024-08-28T01:02:03Z'),
       (2, 'Jane', 'jane@internal.com', true, 'green', '{3, 19}', '{"tags": ["happy"]}', '2024-08-28T01:04:05Z'),
       (3, 'John', 'john@internal.com', false, null, '{}', '{}', '2024-08-28T01:06:07Z'),
       (4, 'Jack', 'jack@internal.com', false, null, null, null, NOW())
;
