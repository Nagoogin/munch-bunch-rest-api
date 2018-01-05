package constants

const USER_TABLE_CREATION_QUERY = `CREATE TABLE IF NOT EXISTS users
(
id SERIAL,
username TEXT NOT NULL,
hash TEXT NOT NULL,
fname TEXT NOT NULL,
lname TEXT NOT NULL,
email TEXT NOT NULL,
hasTruck BOOLEAN NOT NULL,
CONSTRAINT users_pkey PRIMARY KEY (id)
)`

const TRUCK_TABLE_CREATION_QUERY = `CREATE TABLE IF NOT EXISTS trucks
(
id SERIAL,
name TEXT NOT NULL,
CONSTRAINT trucks_pkey PRIMARY KEY (id)
)`

const JWT_SECRET_KEY = "wubbalubbadubdub"

const ERROR = "error"
const SUCCESS = "success"
const NA = "N/A"