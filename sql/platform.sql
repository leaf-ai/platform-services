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
\set quoted_pldb '\'' :pldb '\''

set timezone to 'UTC';

\connect postgres
\set ON_ERROR_STOP on

DO LANGUAGE plpgsql $$
    BEGIN
        IF (SELECT COUNT(*) FROM pg_language WHERE lanname = 'plpgsql') = 0 THEN
            RAISE EXCEPTION 'unsupported sql client variant, use the postgres official client';
        END IF;
    END;
$$;

CREATE FUNCTION pg_temp.CheckDB(dbName text) RETURNS text AS $$
	BEGIN
		IF (SELECT EXISTS (SELECT 1 FROM pg_database WHERE LOWER(datname) = LOWER(dbName)))
			THEN RAISE EXCEPTION 'The database name % supplied matches an existing database', dbName;
		END IF;
		RETURN '';
	END;
$$ LANGUAGE plpgsql;

SELECT pg_temp.CheckDB(:'pldb') INTO TEMP junk;

DROP SCHEMA IF EXISTS public CASCADE;

-- -----------------------------------------------------
-- Create Default DB User, if they are not already present
-- -----------------------------------------------------
DO
$body$
BEGIN
   IF NOT EXISTS (
      SELECT                       -- SELECT list can stay empty for this
      FROM   pg_catalog.pg_user
      WHERE  usename = 'pl') THEN

		CREATE ROLE pl LOGIN
			ENCRYPTED PASSWORD 'md55F4DCC3B5AA765D61D8327DEB882CF99'
			INHERIT CREATEDB CREATEROLE;
   END IF;
END
$body$;

-- -----------------------------------------------------
-- Create Public Schema and assign privs based upon database deployment platform
-- -----------------------------------------------------
DO
$body$
BEGIN
   IF EXISTS (
      SELECT                       -- SELECT list can stay empty for this
      FROM   pg_catalog.pg_roles
      WHERE  rolname = 'postgres') THEN
			CREATE SCHEMA public AUTHORIZATION postgres;
			GRANT postgres TO pl;
   END IF;
   IF EXISTS (
      SELECT                       -- SELECT list can stay empty for this
      FROM   pg_catalog.pg_roles
      WHERE  rolname = 'rds_superuser') THEN
			CREATE SCHEMA public AUTHORIZATION rds_superuser;
			GRANT rds_superuser TO pl;
   END IF;
END
$body$;

CREATE DATABASE :pldb;

\connect :pldb

BEGIN;

GRANT ALL ON DATABASE :pldb TO pl;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO pl;

GRANT ALL ON SCHEMA public TO public;

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

-- Experiments are either active (visible), or deactivated (not visible to traditional users).
-- More states will be added later so this is not a boolean.
--
CREATE TYPE experimentState AS ENUM (
    'Active',
    'Deactivated'
);

CREATE TABLE IF NOT EXISTS experiments (
        id BIGSERIAL,
        uid TEXT NOT NULL,
        state experimentState NOT NULL DEFAULT 'Active',
        created TIMESTAMP NOT NULL DEFAULT 'epoch'::timestamp,
        name TEXT NULL DEFAULT NULL,
        description TEXT NULL DEFAULT NULL
);

CREATE UNIQUE INDEX experiments_uid_idx ON experiments(uid);

-- Layers come in 2 variants, the logic of the enumeration restrictions 
-- are handled inside the db.go code
--
CREATE TYPE layerClass AS ENUM (
    'Input',
    'Output');

-- The following enumeration strings MUST match the strings used by GRPC to map correctly
-- as data is marshalled between SQL and gRPC
CREATE TYPE layerType AS ENUM (
    'Enumeration',
    'Raw',
    'Time',
    'Probability');

-- Constrain the layer numbers
--
CREATE DOMAIN UINT4 AS int4
   CHECK(VALUE >= 0);

CREATE TABLE IF NOT EXISTS layers (
        id BIGSERIAL,
        uid TEXT NOT NULL, -- The unique identifier of the experiment that owns this layer
        number UINT4 NOT NULL, -- The layer number this layer has within the experiment
        name TEXT NOT NULL,
        values TEXT[] NOT NULL,
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
