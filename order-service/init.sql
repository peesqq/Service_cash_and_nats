
CREATE TABLE IF NOT EXISTS "order" (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    title TEXT NOT NULL CHECK( char_length(title) < 32 ),
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

INSERT INTO "order" (user_id, title)
VALUES
    (3, 'order_1'),
    (2, 'order_2'),
    (15, 'order_3'),
    (7, 'order_4'),
    (4, 'order_5'),
    (5, 'order_6');
