-- The schema definition for the platform service plane containing tables
-- user across multiple services, for now.
--
-- This file is compatible with the Aurora RDS account and security naming conventions
-- for roles etc.  To run this file to generate a new DB you can use a command 
-- such as the following:
--
-- PGUSER=pl PGHOST=dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com PGDATABASE=platform psql -f platform.sql
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
        ENCRYPTED PASSWORD 'md55F4DCC3B5AA765D61D8327DEB882CF99'
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
        id BIGSERIAL,
        uid TEXT NOT NULL,
        created TIMESTAMP NOT NULL DEFAULT 'epoch'::timestamp,
        name TEXT NULL DEFAULT NULL,
        description TEXT NULL DEFAULT NULL
);

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
        id BIGSERIAL,
        uid TEXT NOT NULL, -- The unique identifier of the experiment that owns this layer
        number INT NOT NULL, -- The layer number this layer has within the experiment
        name TEXT NOT NULL,
        class layerClass NOT NULL,
        type layerType NOT NULL
);

CREATE UNIQUE INDEX layers_idx ON layers(uid, number);

-- -----------------------------------------------------
-- Alter Table Starting Sequence Numbers to avoid test case number ranges etc
-- -----------------------------------------------------

ALTER SEQUENCE experiments_id_seq RESTART WITH 1000 OWNED BY experiments.id;
ALTER SEQUENCE layers_id_seq RESTART WITH 1000 OWNED BY layers.id;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO pl;

INSERT INTO upgrades (version) VALUES (1);

COMMIT;
