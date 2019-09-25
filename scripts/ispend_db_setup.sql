/*
one to many example
https://launchschool.com/books/sql/read/table_relationships#manytomany
*/

DROP TABLE IF EXISTS spends;
DROP TABLE IF EXISTS spend_kinds;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id serial UNIQUE NOT NULL,
    email varchar(35) UNIQUE,
    username varchar(35) UNIQUE NOT NULL,
    password varchar(35) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE spend_kinds (
    id serial UNIQUE NOT NULL,
    user_id integer NOT NULL,
    name varchar(35) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE spends (
    id serial UNIQUE NOT NULL,
    currency char(10) NOT NULL,
    amount integer NOT NULL,
    timestamp timestamp NOT NULL, /*DEFAULT CURRENT_TIMESTAMP,*/
    user_id integer NOT NULL,
    kind_id integer NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (kind_id) REFERENCES spend_kinds(id) ON DELETE CASCADE
);

/*
Email      string      `json:"email"`
Username   string      `json:"username"`
Password   string      `json:"password"`
Spends     []Spending  `json:"spends"`
SpendKinds []SpendKind `json:"spending_kinds"`

Currency  string     `json:"currency"`
Amount    float32    `json:"amount"`
Kind      *SpendKind `json:"kind"`
Timestamp time.Time  `json:"timestamp"`

Name  string     `json:"name"`
*/
