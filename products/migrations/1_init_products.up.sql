CREATE TABLE IF NOT EXISTS products
    (
        uid serial PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT NOT NULL,
        price double precision NOT NULL,
        delete BOOLEAN NOT NULL DEFAULT false,
        quantity INT NOT NULL
    );

CREATE TABLE IF NOT EXISTS purchases
    (
        uid serial PRIMARY KEY,
        user_id INT NOT NULL,
        product_id INT NOT NULL REFERENCES products(uid),
        quantity INT NOT NULL,
        purchase_date TIMESTAMP NOT NULL DEFAULT NOW()
    );