DROP TABLE IF EXISTS spends;
DROP TABLE IF EXISTS spend_kinds;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS default_spend_kinds;

CREATE TABLE users (
    id serial PRIMARY KEY,
    email varchar(35) UNIQUE,
    username varchar(35) UNIQUE NOT NULL,
    password varchar(35) NOT NULL
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
    amount integer NOT NULL,
    timestamp timestamp NOT NULL, /*DEFAULT CURRENT_TIMESTAMP,*/
    user_id integer NOT NULL,
    kind_id integer NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (kind_id) REFERENCES spend_kinds(id) ON DELETE CASCADE
);

INSERT INTO default_spend_kinds (name) VALUES ('Travel');
INSERT INTO default_spend_kinds (name) VALUES ('Nightlife');
INSERT INTO default_spend_kinds (name) VALUES ('Rent');
INSERT INTO default_spend_kinds (name) VALUES ('Food');

INSERT INTO users (email, username, password) VALUES ('admin@serjspends.de', 'admin', 'admin1');
INSERT INTO users (email, username, password) VALUES ('lazar@serjspends.de', 'lazar', 'lazar1');
