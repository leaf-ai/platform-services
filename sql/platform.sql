-- The schema definition for the platform service plane containing tables
-- user across multiple services, for now.
--
\set pldb `echo "$PGDATABASE"`

set timezone to 'UTC';

\connect postgres

DROP SCHEMA IF EXISTS public CASCADE;
DROP DATABASE IF EXISTS :pldb;

-- -----------------------------------------------------
-- Create Default DB User
-- -----------------------------------------------------
DROP ROLE IF EXISTS pl;
CREATE ROLE pl LOGIN
        ENCRYPTED PASSWORD 'md5e270fe4980444975ebe8138bdfcce914'
        INHERIT CREATEDB CREATEROLE;

-- -----------------------------------------------------
-- Create Public Schema
-- -----------------------------------------------------

DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public AUTHORIZATION postgres;

CREATE DATABASE :pldb;

\connect :pldb

BEGIN;

GRANT ALL ON DATABASE :pldb TO pl;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO pl;

GRANT ALL ON SCHEMA public TO public;

GRANT rds_superuser TO pl;

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN schema public TO pl;

-- -----------------------------------------------------
-- Create All Sequences
-- -----------------------------------------------------

CREATE SEQUENCE experiment_id_seq;
CREATE SEQUENCE layer_id_seq;

-- -----------------------------------------------------
-- Create All Tables
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS upgrades (
        version INT NOT NULL,
        timestamp TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS logs (
        priority INT NOT NULL,
        timestamp TIMESTAMP NOT NULL,
        source VARCHAR(128) NOT NULL,
        msg VARCHAR(128) NOT NULL
);

CREATE INDEX logs_source_timestamp_idx ON logs (source, timestamp);

CREATE TABLE IF NOT EXISTS experiments (
        id BIGINT NOT NULL DEFAULT nextval('experiment_id_seq'),
        uid TEXT NOT NULL,
        created TIMESTAMP NOT NULL DEFAULT 'epoch'::timestamp,
        name TEXT NULL DEFAULT NULL,
        description TEXT NULL DEFAULT NULL,
        PRIMARY KEY (id) -- Database generated
);

CREATE UNIQUE INDEX experiments_idx ON experiments(id);
CREATE UNIQUE INDEX experiments_uid_idx ON experiments(uid);

CREATE TYPE layerClass AS ENUM (
    'input',
    'output');

CREATE TYPE layerType AS ENUM (
    'enum',
    'raw',
    'time',
    'probability');

CREATE TABLE IF NOT EXISTS layers (
        id BIGINT NOT NULL DEFAULT nextval('layer_id_seq'),
        name TEXT NOT NULL,
        class layerClass NOT NULL,
        type layerType NOT NULL,
        PRIMARY KEY (id) -- Database generated
);

CREATE UNIQUE INDEX layers_idx ON layers(id);

-- -----------------------------------------------------
-- Grant All Permissions
-- -----------------------------------------------------

GRANT USAGE, SELECT ON SEQUENCE experiment_id_seq TO pl;
GRANT USAGE, SELECT ON SEQUENCE layer_id_seq TO pl;

-- -----------------------------------------------------
-- Alter Table Starting Sequence Number
-- -----------------------------------------------------

ALTER SEQUENCE experiment_id_seq RESTART WITH 1000;
ALTER SEQUENCE layer_id_seq RESTART WITH 1000;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO pl;

INSERT INTO upgrades (version) VALUES (1);

COMMIT;
