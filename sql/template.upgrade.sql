-- ------------------------------------------------------
-- A template for performing updates to a Database
-- schema initially checking the version followed by version 
-- updates to the Database associated parameters on success
-- ------------------------------------------------------

\set ON_ERROR_STOP on
\set pldb `echo "$PGDATABASE"`

set timezone to 'UTC';

\connect :pldb


BEGIN;

-- The following code block will validate the existing version of the Database.  Please
-- update the version number inside this block to reflect your current version

DO $do$
BEGIN
    IF (SELECT version FROM upgrades ORDER BY timestamp DESC LIMIT 1) <> '1' THEN
        RAISE EXCEPTION 'The wrong version of Database is being used for this upgrade';
    END IF ;
END;
$do$ language plpgsql;

-- Place your schema modification code here

-- Update the integer to reflect the version being stepped to in the following three statements

INSERT INTO upgrades (version) VALUES (2);

COMMIT;
