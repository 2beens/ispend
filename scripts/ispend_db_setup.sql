SET TIME ZONE '+00:00';

DROP TABLE IF EXISTS spends;
DROP TABLE IF EXISTS spend_kinds;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS default_spend_kinds;

CREATE TABLE users (
    id serial PRIMARY KEY,
    email varchar(35) UNIQUE,
    username varchar(35) UNIQUE NOT NULL,
    password varchar(130) NOT NULL
);

CREATE TABLE spend_kinds (
    id serial PRIMARY KEY,
    user_id integer NOT NULL,
    name varchar(35) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE default_spend_kinds (
    id serial PRIMARY KEY,
    name varchar(35) UNIQUE NOT NULL
);

CREATE TABLE spends (
    id serial PRIMARY KEY,
    currency char(10) NOT NULL,
    amount real NOT NULL,
    spend_timestamp timestamp NOT NULL, /*DEFAULT CURRENT_TIMESTAMP,*/
    user_id integer NOT NULL,
    kind_id integer NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (kind_id) REFERENCES spend_kinds(id) -- ON DELETE CASCADE
);

INSERT INTO default_spend_kinds (name) VALUES ('Travel');
INSERT INTO default_spend_kinds (name) VALUES ('Nightlife');
INSERT INTO default_spend_kinds (name) VALUES ('Rent');
INSERT INTO default_spend_kinds (name) VALUES ('Food');

INSERT INTO users (email, username, password) VALUES ('admin@serjspends.de', 'admin', '$2a$14$OhXgytZEMIxgFr9q02cyru72BJwFZ3zMWEf92/YvinjzmFhYeyfLS');
INSERT INTO users (email, username, password) VALUES ('lazar@serjspends.de', 'lazar', '$2a$14$OhXgytZEMIxgFr9q02cyru72BJwFZ3zMWEf92/YvinjzmFhYeyfLS');

INSERT INTO spend_kinds (user_id, name) VALUES (1, 'House');
INSERT INTO spend_kinds (user_id, name) VALUES (1, 'Car');
INSERT INTO spend_kinds (user_id, name) VALUES (2, 'House');
INSERT INTO spend_kinds (user_id, name) VALUES (2, 'Car');

INSERT INTO spends (currency, amount, spend_timestamp, user_id, kind_id) VALUES ('EUR', 1000, '2019-02-02 00:00:01', 1, 2);
INSERT INTO spends (currency, amount, spend_timestamp, user_id, kind_id) VALUES ('RSD', 120.60, '2019-01-01 00:00:01', 2, 2);