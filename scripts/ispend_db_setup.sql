DROP TABLE IF EXISTS spends;
DROP TABLE IF EXISTS spend_kinds;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS default_spend_kinds;

CREATE TABLE users (
    id serial UNIQUE NOT NULL,
    email varchar(35) UNIQUE,
    username varchar(35) UNIQUE NOT NULL,
    password varchar(35) NOT NULL,
    PRIMARY KEY (id)    /* -> TODO: is this line redundant, becaues of "id serial" ? */
);

CREATE TABLE spend_kinds (
    id serial UNIQUE NOT NULL,
    user_id integer NOT NULL,
    name varchar(35) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE TABLE default_spend_kinds (
    id serial UNIQUE NOT NULL,
    name varchar(35) UNIQUE NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE spends (
    id serial UNIQUE NOT NULL,
    currency char(10) NOT NULL,
    amount integer NOT NULL,
    timestamp timestamp NOT NULL, /*DEFAULT CURRENT_TIMESTAMP,*/
    user_id integer NOT NULL,
    kind_id integer NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (kind_id) REFERENCES spend_kinds(id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

INSERT INTO default_spend_kinds VALUES (1, 'Travel');
INSERT INTO default_spend_kinds VALUES (2, 'Nightlife');
INSERT INTO default_spend_kinds VALUES (3, 'Rent');
INSERT INTO default_spend_kinds VALUES (4, 'Food');

INSERT INTO users VALUES (1, 'admin@serjspends.de', 'admin', 'admin1');
INSERT INTO users VALUES (2, 'lazar@serjspends.de', 'lazar', 'lazar1');
