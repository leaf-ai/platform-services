package db

// The db.go file contains a Go lang marshalling layer on top of the Postgres and
// Go SQL layers.  One of the responsibilities of the db.go module is also to
// enable Postgres arrays and the like to be marshalled into Thrift

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/davecgh/go-spew/spew"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	sq "github.com/Masterminds/squirrel"
)

const ExpectedDBVersion int64 = 1

var (
	databaseHostPort    = flag.String("dbaddr", defaultDBHostPort(), "Host:Port pair for the database server")
	databaseUser        = flag.String("dbuser", defaultDBUser(), "User name for accessing the database")
	databaseName        = flag.String("dbname", defaultDBName(), "The 'DarkCycle' database name")
	databaseOptions     = flag.String("dboptions", "", "Postgres options for inclusion in the 'DarkCycle' DSN, for example -dboptions=options=\"-c statement_timeout=2min\"")
	databaseMaxConn     = flag.Int("dbmaxconn", 72, "Sets a limit for open connections the master will use to postgres")
	databaseMaxIdleConn = flag.Int("dbmaxidleconn", 8, "Sets a limit for open and idle connections the master will use to postgres")
)

const defaultDBPassword string = "NjlmdGVsdmlz"
const defaultPassHash string = "688c444a191cd8220b9ce797b06d74a9"

func defaultDBUser() string {
	name := os.Getenv("PGUSER")
	if name == "" {
		return "dc"
	}
	return name
}

func defaultDBName() string {
	name := os.Getenv("PGDATABASE")
	if name == "" {
		return "darkcycle"
	}
	return name
}

func defaultDBHostPort() string {

	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")

	if host == "" {
		host = "127.0.0.1"
	}

	if port == "" {
		port = "5432"
	}

	return fmt.Sprintf("%s:%s", host, port)
}

var dBase *sqlx.DB

// Log is used to hold unstructured log entries that will be marshalled
//
type Log struct {
	Priority    syslog.Priority
	Nanoseconds int64
	ComponentID string
	Msg         string
}

// ExperimentData will hold data ready for being handled by the SQL layer for experiments.
//
type ExperimentData struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Created     time.Time `db:"created"`
}

// dbDownErr is used within the db layer to record if the DB is down.  A simple circuit breaker is
// used with the DB gofunc handler loop to retry connections on a regular basis.  The DB methods
// will respect this flag and fail out any calls until the circuit breaker
//
var dbDownErr = &dclib.SafeError{}

func dbHasCorrectVersion(expect int64) (err error, ok bool, actual int64) {
	err = dBase.QueryRow("SELECT version FROM upgrades ORDER BY timestamp DESC LIMIT 1").Scan(&actual)

	return err, (expect == actual), actual
}

// CloseDB is used to close any existing database connections to the DarkCycle DB
//
func CloseDB() error {

	message := "database has been closed intentionally"
	logW.Trace(message)

	defer dbDownErr.Set(fmt.Errorf(message))
	return dBase.Close()
}

// getPassFromPGPass matches the very first line that can be matched from the postgres password file as documented
// by https://www.postgresql.org/docs/9.5/static/libpq-pgpass.html
//
func getPassFromPGPass(passFile string, host string, port string, db string, user string) (pass string, err error) {

	logW.Debug(fmt.Sprintf("%s", passFile), "host", host, "port", port, "database", db, "user", user)
	file, err := os.Open(passFile)
	if err != nil {
		logW.Error(fmt.Sprintf("Unable to open the postgres password file '%s' due to %s", passFile, err.Error()), "error", err)
		return pass, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cred := strings.Split(scanner.Text(), ":")
		// Make sure we have at least five fields in the list with naieve splits
		if len(cred) < 5 {
			continue
		}

		// After an initial split go through looking for backslashes and joining tokens together to form
		// the full unescaped fields if a backslash is seen
		//
		tokens := []string{}
		continued := false
		for _, field := range cred {
			if continued {
				tokens[len(tokens)-1] = tokens[len(tokens)-1] + field
			} else {
				tokens = append(tokens, field)
			}
			continued = strings.HasSuffix(field, "\\")
		}

		// Now after bringing the fields together after processing
		// escape characters check to make sure we have the exact number of fields expected
		if len(tokens) != 5 {
			continue
		}

		if (tokens[0] != "*" && tokens[0] != host) ||
			(tokens[1] != "*" && tokens[1] != port) ||
			(tokens[2] != "*" && tokens[2] != db) ||
			(tokens[3] != "*" && tokens[3] != user) {
			continue
		}
		return tokens[4], nil
	}

	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return "", err
}

// StartDB is used to open and then attempt to test the connection to the
// DarkCycle database, which contains state information for components of the
// DarkCycle ecosystem
//
func StartDB(quitC <-chan bool) (err error) {
	// The following function does not create a Live database connection.  This is done later
	// during the normal server life cycle
	//
	if err := initDarkcycleDB(*databaseHostPort, *databaseUser); err != nil {
		msg := fmt.Sprint("Could not initialize the Darkcycle database at ", *databaseHostPort, " due to ", err.Error())
		logW.Error(msg, "dbHostPort", *databaseHostPort, "dbUser", *databaseUser, "error", err)
		return err
	}

	// On the first time the master is started the DB is left in a down state and is initialized
	// during the normal life cycle of the master server.  This is done as a requirement of DC-1002.
	//
	dbDownErr.Set(fmt.Errorf("Darkcycle database start not yet completed"))

	// Start a runtime monitor for DB connections held open, a liveness check
	// and a circuit breaker
	//
	go func() {

		defer CloseDB()

		// Start by instantly trying to get the database up and going, this duration
		// will be reset once the main service loop has started below to a much
		// longer interval that is appropriate for casually checking if the DB
		// is live.
		//
		dbCheckTimer := time.Duration(time.Microsecond)

		for {
			select {
			case <-time.After(dbCheckTimer):
				dbRecovery := !dbDownErr.IsErr()
				if err := dBase.Ping(); err != nil {

					dbDownErr.Set(err)

					msg := fmt.Sprint("Darkcycle database ", *databaseHostPort, " ", databaseName, " is currently down due to ", err.Error())
					logW.Warning(msg, "dbHostPort", *databaseHostPort, "dbName", databaseName, "error", err)
					dbCheckTimer = time.Duration(5 * time.Second)
					continue
				}

				if dbRecovery {
					logW.Info("DarkCycle database startup / recovery has been performed ", "dbHostPort", *databaseHostPort, "dbName", databaseName)

					// Check that the Database has the expected version
					err, ok, version := dbHasCorrectVersion(ExpectedDBVersion)
					if err == nil && !ok {
						msg := fmt.Sprint("DarkCycle DB has the wrong version ", ExpectedDBVersion, " expected got version ", version, " instead")
						logW.Fatal(msg, "dbExpectedVersion", ExpectedDBVersion, "dbVersion", version)
						os.Exit(-4)
					}
					if err != nil {
						logW.Fatal("DB has no version marker, suspect or incorrect database schema", "error", err)
						os.Exit(-5)
					}
				}

				dbDownErr.Clear()
				msg := fmt.Sprint("Darkcycle database has ", dBase.Stats().OpenConnections, " connections")
				logW.Info(msg, "dbHostPort", *databaseHostPort, "dbName", *databaseName, "dbConnectionCount", dBase.Stats().OpenConnections)

			case <-quitC:
				logW.Debug("Darkcycle database monitor stopped")
				return
			}

			dbCheckTimer = time.Duration(15 * time.Second)
		}
	}()

	return nil
}

// Status is used to retrieve the main DarkCycle eco system status from a master
//
func GetDBStatus() (status *thriftdci.DarkCycleStatus, err error) {

	// Prefetch the Work Unit Server errors so that we can alloce the array inside the
	// status report
	//
	// TODO wuFailures := FetchWURecentFailures()
	//
	// Initialize the datastructure for reporting with blank data
	status = &thriftdci.DarkCycleStatus{
		State:    thriftdci.EntityState_RUNNING,
		SqlState: thriftdci.EntityState_RUNNING,
		SqlError: "",
	}

	// Load any darkcycle DB failure information
	if err = dbDownErr.Get(); err != nil {
		status.SqlState = thriftdci.EntityState_ERROR
		status.SqlError = dbDownErr.String()
	}

	return status, nil
}

// initDarkcycleDB will setup the database configuration but does not actually create a live connection
// or validate the parameters supplied to it.
//
func initDarkcycleDB(url string, user string) (err error) {

	pgPassFile := os.Getenv("PGPASSFILE")
	if len(pgPassFile) == 0 {
		if pgPassHome := os.Getenv("HOME"); len(pgPassHome) != 0 {
			pgPassFile = filepath.Join(pgPassHome, ".pgpass")
		}
		if _, err = os.Stat(pgPassFile); os.IsNotExist(err) {
			pgPassFile = "~/.pgpass"
		}
		// If standard sensible locations dont work due to
		// shell issues in the salt startup try a hard coded but
		// well known location
		if _, err = os.Stat(pgPassFile); os.IsNotExist(err) {
			pgPassFile = "/home/darkcycle/.pgpass"
		}
	}

	hostPort := strings.Split(url, ":")
	password, err := getPassFromPGPass(pgPassFile, hostPort[0], hostPort[1], *databaseName, user)

	dbOptions := ""

	if 0 != len(*databaseOptions) {
		dbOptions = fmt.Sprintf("&%s", *databaseOptions)
	}

	datasetName := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable%s", user, password, url, *databaseName, dbOptions)
	datasetName = strings.Replace(datasetName, ":@", "@", 1)
	safeDatasetName := fmt.Sprintf("postgres://%s:********@%s/%s?sslmode=disable%s", user, url, *databaseName, dbOptions)

	logW.Debug(fmt.Sprint("Connecting to Darkcycle database ", safeDatasetName), "dbConnectString", safeDatasetName)
	db, err := sql.Open("postgres", datasetName)
	if err != nil {
		logW.Error(fmt.Sprint("Could not open postgres Darkcycle database ", safeDatasetName), "dbConnectString", safeDatasetName, "error", err)
		return err
	}
	logW.Info(fmt.Sprint("Darkcycle database configured ", safeDatasetName), "dbConnectString", safeDatasetName)

	db.SetMaxOpenConns(*databaseMaxConn)
	db.SetMaxIdleConns(*databaseMaxIdleConn)

	dBase = sqlx.NewDb(db, "postgres")
	// The follow functor takes a Thrift style name and converts it to a DB column style name
	//
	// This method is temporary until the golang sql treats arrays as first class citizens.
	// When that happens we will be able to remove all of our local marshalling code and
	// structures within this module
	dBase.MapperFunc(func(input string) string {

		var output bytes.Buffer
		wasUpper := false
		for _, aRune := range input {
			wasUpper = unicode.IsUpper(aRune)
			break
		}

		for i, aRune := range input {
			if i != 0 && !wasUpper && unicode.IsUpper(aRune) {
				output.WriteString("_")
			}
			output.WriteRune(unicode.ToLower(aRune))

			wasUpper = unicode.IsUpper(aRune)
		}
		return output.String()
	})

	return nil
}

// showAllTrace is a utility function useful for when the database comes up to dump information
// about the session and other parameters that can be useful in debugging issues with parameters
// set by the operational side of the system.  Its output is generally intended for engineering
// and thrid level support personnel
//
func DBShowAllTrace() {

	rows, err := dBase.Queryx("show all")

	if err != nil {
		logW.Error(fmt.Sprint("Postgres show all failed due to ", err.Error()), "error", err)
		return
	}
	defer rows.Close()

	type DBSetting struct {
		Name        string
		Setting     string
		Description string
	}

	for rows.Next() {
		aRow := &DBSetting{}
		if err = rows.StructScan(aRow); err != nil {
			logW.Error(fmt.Sprint("DB settings not available due to ", err.Error()), "error", err)
			return
		}
		msg := fmt.Sprintf("%s='%s' %s", aRow.Name, aRow.Setting, aRow.Description)
		logW.Trace(fmt.Sprint("DB Setting ", msg), aRow.Name, msg)
	}
}

// CheckIfFatal is used to handle the errors being returned from the Postgres driver.  If
// errors exist that are not recoverable that relate to schema issues, such as missing columns,
// indexes and generally expected Database artifacts then this function will kill the task in which
// they are running.
//
func CheckIfFatal(inErr error) (err error) {
	if inErr != nil {
		if strings.HasPrefix(inErr.Error(), "sql:") {

			switch inErr {
			case sql.ErrTxDone, sql.ErrNoRows:
				return inErr
			}

			logW.Fatal(fmt.Sprint("golang SQL failure return code found ", inErr.Error()), "error", inErr)

			os.Exit(-5)
		}

		if driverErr, ok := inErr.(*pq.Error); ok {

			logW.Warning(fmt.Sprint("postgres SQL return code ", driverErr.Error()), "dbErrorCode", driverErr.Code.Name, "dbErrorClass", driverErr.Code.Class())
			classNumber, _ := strconv.Atoi(string(driverErr.Code.Class()))

			// Look for classes of errors that can be returned to the caller which are not Fatal.
			//
			// For further information please read http://www.postgresql.org/docs/9.5/static/errcodes-appendix.html
			//
			switch classNumber {
			case 0:
				// Class 00 — Successful Completion
				// 00000 successful_completion
				return nil
			case 8:
				// Class 08 — Connection Exception
				//
				// 08000 connection_exception
				// 08003 connection_does_not_exist
				// 08006 connection_failure
				// 08001 sqlclient_unable_to_establish_sqlconnection
				// 08004 sqlserver_rejected_establishment_of_sqlconnection
				// 08007 transaction_resolution_unknown
				// 08P01 protocol_violation
				return inErr
			case 23:
				// Class 23 — Integrity Constraint Violation
				//
				// 23000 integrity_constraint_violation
				// 23001 restrict_violation
				// 23502 not_null_violation
				// 23503 foreign_key_violation
				// 23505 unique_violation
				// 23514 check_violation
				// 23P01 exclusion_violation
				return inErr
			}

			// Now the error infomation is accessible for any other classes of errors
			// check those
			switch driverErr.Code.Name() {
			case "successful_completion":
				return nil
			default:
				logW.Fatal(fmt.Sprint("Postgres driver fatal error ", driverErr.Error()), "dbDriverError", driverErr, "dbErrorCode", driverErr.Code.Name())

				os.Exit((classNumber + 10) * -1)
			}
			return nil
		}
	}

	return inErr
}

// SelectProject is used to retrieve one or more postgres DB rows and
// marshall these into corresponding Thrift compatible data structures that
// are used by the rest of the masters software components.
//
// If the id is specified, that is not 0, then the owner will not be used
// for selecting which record is returned.  If the id is left to be a zero
// value then the owner will be used to select which rows are returned.
//
// The function can return both a nil, or an empty array for the first result,
// along with a nil for the error in the case the SQL query works but returns
// no data.
//
func SelectProject(userId int64, orgId int64, id int64) (result []*thriftdci.Project, err error) {

	if err = dbDownErr.Get(); err != nil {
		return nil, err
	}

	logW.Dump("userId", userId, "organizationId", orgId, "projectId", id)

	query := sq.Select("id,state,name,description,org_id,app_id,work_unit_servers,dataset_ids,service_ids,started_at,started_with,last_sample_at,last_sample_count,sum_times_sqr_secs,sum_times_secs").
		From("projects").
		OrderBy("id DESC").
		PlaceholderFormat(sq.Dollar)

	useAnd := false
	if id > 0 {
		query = query.Where(sq.Eq{"id": id})
		useAnd = true
	}

	if orgId > 0 {
		if useAnd {
			query = query.Where(sq.And{sq.Eq{"org_id": orgId}})
		} else {
			query = query.Where(sq.Eq{"org_id": orgId})
		}
		useAnd = true
	}

	if userId > 0 {
		err = fmt.Errorf("selecting projects using the user ID is not yet supported, please use a 0 for the userId parameter")
		logW.Warning(err.Error(), "userId", userId, "organizationId", orgId, "projectId", id, "error", err)
		return nil, err
	}

	if !useAnd {
		err = fmt.Errorf("selecting projects requires at least 1 value to specified as input")
		logW.Warning(err.Error(), "error", err)
		return nil, err
	}

	sql, args, err := query.ToSql()
	if CheckIfFatal(err) != nil {
		logW.Error(fmt.Sprint("Unable to build query select project due to ", err.Error()), "userId", userId, "organizationId", orgId, "projectId", id, "error", err)
		return nil, err
	}

	rows, err := dBase.Queryx(sql, args...)
	if CheckIfFatal(err) != nil {
		logW.Error(fmt.Sprint("Unable to execute query select project due to ", err.Error()), "userId", userId, "organizationId", orgId, "projectId", id, "error", err)
		return nil, err
	}
	defer rows.Close()

	result = []*thriftdci.Project{}
	for rows.Next() {
		row := ProjectData{}
		if err = rows.StructScan(&row); CheckIfFatal(err) != nil {
			logW.Error(fmt.Sprint("Unable to retrieve select projects due to ", err.Error()), "userId", userId, "organizationId", orgId, "projectId", id, "error", err)
			return nil, err
		}

		startedAt := row.StartedAt.UnixNano()
		lastSampleAt := row.LastSampleAt.UnixNano()

		trow := &thriftdci.Project{
			ID:              row.ID,
			State:           row.State,
			Name:            row.Name,
			Description:     row.Description,
			OrgID:           row.OrgID,
			AppID:           row.AppID,
			WorkUnitServers: pgToArrayStr(row.WorkUnitServers),
			DatasetIds:      pgToInt64(row.DatasetIds),
			ServiceIds:      pgToInt64(row.ServiceIds),
			StartedAt:       &startedAt,
			StartedWith:     &row.StartedWith,
			LastSampleAt:    &lastSampleAt,
			LastSampleCount: &row.LastSampleCount,
			SumTimesSqrSecs: &row.SumTimesSqrSecs,
			SumTimesSecs:    &row.SumTimesSecs,
		}
		result = append(result, trow)
	}

	logW.Dump("userId", userId, "organizationId", orgId, "projectId", id, "projects", result)

	return result, nil
}

// InsertProject is used to insert a single dataset record into the
// masters postgres database. A unique ID will be generated by the insertion
// operation and this ID will be placed into the returned record so that the
// caller has a unique reference to the inserted data for use in either
// performing other database operations or making reference to the project
// from other records.
//
// The result returned this function does not share memory with the input data
// parameter and is a deep copy.  This is slight less efficent than returning the
// mutated input structure but far safer in regards to potential side effects as a
// result of assumptions made by the caller.
//
//
func InsertProject(data *thriftdci.Project) (result *thriftdci.Project, err error) {

	if err = dbDownErr.Get(); err != nil {
		return nil, err
	}

	logW.Dump("project", data)

	// The output result is a DeepCopy of the input with the ID replaced to prevent side effects
	// caused by mutating the input
	//
	result = ProjectDeepCopy(data)

	query := sq.Insert("projects").
		Columns("state", "name", "description", "org_id", "app_id", "work_unit_servers", "dataset_ids", "service_ids").
		Values(data.State, data.Name, data.Description, data.OrgID, data.AppID, arrayStrToPg(data.WorkUnitServers), arrayInt64ToPg(data.DatasetIds), arrayInt64ToPg(data.ServiceIds)).
		Suffix("RETURNING \"id\"").
		RunWith(dBase.DB).
		PlaceholderFormat(sq.Dollar)

	if err = query.QueryRow().Scan(&result.ID); CheckIfFatal(err) != nil {
		logW.Error(fmt.Sprint("Unable to retrieve insert project due to ", err.Error()), "project", data, "error", err)
		return nil, err
	}
	return result, nil
}

// UpdateProject can be used to modify the contents of an existing database row.  The
// ID found within the input data record is used to identify the DB row to be replaced
// or updated with the new data in place.
//
// If the intent is to only change the single state item inside the application row then
// the ChangeProject function should be used.
//
// The result returned this function does not share memory with the input data
// parameter and is a deep copy.  This is slight less efficent than returning the
// mutated input structure but far safer in regards to potential side effects as a
// result of assumptions made by the caller.
//
// TODO : DC-1014 : Database : ChangeProject and other state changes need constraints
//
func UpdateProject(data *thriftdci.Project) (result *thriftdci.Project, err error) {

	if err = dbDownErr.Get(); err != nil {
		return nil, err
	}

	// The output result is a DeepCopy of the input with the ID replaced to prevent side effects
	// caused by mutating the input
	//
	result = ProjectDeepCopy(data)

	query := sq.
		Update("projects").
		Set("state", data.State).
		Set("name", data.Name).
		Set("description", data.Description).
		Set("org_id", data.OrgID).
		Set("app_id", data.AppID).
		Set("work_unit_servers", arrayStrToPg(data.WorkUnitServers)).
		Set("dataset_ids", arrayInt64ToPg(data.DatasetIds)).
		Set("service_ids", arrayInt64ToPg(data.ServiceIds)).
		Where("state IN (?,?) ", thriftdci.EntityState_CREATED, thriftdci.EntityState_INVALID).
		Where(sq.And{sq.Eq{"id": data.ID}}).
		RunWith(dBase.DB).
		Suffix(`RETURNING "id"`).
		PlaceholderFormat(sq.Dollar)

	sqlResult, err := query.Exec()
	if CheckIfFatal(err) != nil {
		logW.Error(fmt.Sprint("Unable to execute update project query due to ", err.Error()), "project", data, "error", err)
		return nil, err
	}

	if count, err := sqlResult.RowsAffected(); count != 1 {
		if err != nil {
			err = fmt.Errorf("UpdateProject was not able to modify data due to %s", err)
		} else {
			err = fmt.Errorf("UpdateProject was not able to modify data using the supplied input %d. The DB entity state may have been invalid.", data.ID)
		}
		logW.Warning(fmt.Sprint("UpdateProject failed due to ", err.Error()), "project", data, "rowCount", count, "error", err)
		return nil, err
	}

	logW.Dump("project", result)

	return result, nil
}

// ChangeProject can be used to modify the application specific state of an
// existing database row.  The ID found within the input data record is used to
// identify the DB row to be replaced or updated with the new data in place.
//
// TODO : DC-1014 : Database : ChangeProject and other state changes need constraints
//
func ChangeProject(id int64, state thriftdci.EntityState) (err error) {

	if err = dbDownErr.Get(); err != nil {
		return err
	}

	if state == thriftdci.EntityState_RUNNING {
		err = fmt.Errorf("projects (%d) cannot be changed to the running state without using the UpdateProjectStarted function", id)
		logW.Error(fmt.Sprint("ChangeProject failed due to ", err.Error()), "projectId", id, "state", state)
		return err
	}

	logW.Dump("projectId", id, "state", state)

	update := sq.
		Update("projects").
		Set("state", state).
		Where(sq.Eq{"id": id}).
		RunWith(dBase.DB).
		Suffix(`RETURNING "id"`).
		PlaceholderFormat(sq.Dollar)

	if _, err := update.Exec(); CheckIfFatal(err) != nil {
		logW.Error(fmt.Sprint("Unable to execute change project due to ", err.Error()), "projectId", id, "state", state, "error", err)
		return err
	}
	return nil
}

// SaveLog is used to insert a logging record either originating from the master,
// or from a thrift client into the masters logging DB
//
func SaveLog(log *Log) (err error) {

	if err = dbDownErr.Get(); err != nil {
		return err
	}

	message := log.Msg
	if len(message) > 127 {
		message = message[:127]
	}

	insert := sq.Insert("logs").
		Columns("priority", "timestamp", "component_id", "msg", "project_id", "workunit_id").
		Values(int32(log.Priority), time.Unix(0, log.Nanoseconds), log.ComponentID, message, log.ProjectID, log.WorkUnitID).
		RunWith(dBase.DB).
		PlaceholderFormat(sq.Dollar)

	if _, err = insert.Exec(); CheckIfFatal(err) != nil {
		logW.Error(fmt.Sprint("log record could not be written to the DB due to ", err.Error()), "error", err)
		return err
	}
	return nil
}

// Pack Golang arrays of 64 Bit ints into Uint8 structures used by Postgres
//
func arrayInt64ToPg(a2 []int64) (result []uint8) {
	s2 := "{"
	for _, v2 := range a2 {
		if s2 != "{" {
			s2 += ","
		}
		s2 += strconv.Itoa(int(v2))
	}
	s2 += "}"
	result = make([]byte, len(s2))
	for i := range s2 {
		result[i] = s2[i]
	}
	return result
}

// Pack Golang arrays of 64 Bit ints into Uint8 structures used by Postgres
//
func pgToInt64(a []uint8) (result []int64) {
	result = []int64{}
	if len(a) <= 2 {
		return result
	}
	c1 := string([]byte(a[:]))
	c2 := strings.Replace(c1, "{", "", -1)
	c3 := strings.Replace(c2, "}", "", -1)
	c4 := strings.Split(c3, ",")
	for _, v := range c4 {
		n, _ := strconv.Atoi(v)
		result = append(result, int64(n))
	}
	return result
}

// Pack Uint8 structures used by Postgres into Golang arrays of strings
//
func pgToArrayStr(arg1 []uint8) (result []string) {

	result = []string{}

	arg2 := string([]byte(arg1[:]))
	// Completely empty array test
	if arg2 != "{}" {
		arg21 := strings.Replace(arg2, "{", "", -1)
		arg22 := strings.Replace(arg21, "}", "", -1)

		lastQuote := rune(0)
		f := func(c rune) bool {
			switch {
			case c == lastQuote:
				lastQuote = rune(0)
				return false
			case lastQuote != rune(0):
				return false
			case unicode.In(c, unicode.Quotation_Mark):
				lastQuote = c
				return false
			default:
				return c == ','

			}
		}

		m := strings.FieldsFunc(string(arg22), f)

		for _, item := range m {
			unescaped := string(item)
			for _, aRune := range unescaped {
				if unicode.Is(unicode.Quotation_Mark, aRune) {
					last, _ := utf8.DecodeLastRuneInString(unescaped)
					if aRune == last {
						unescaped = unescaped[1 : len(unescaped)-1]
					}
				}
				break
			}
			result = append(result, unescaped)
		}
	}
	return result
}

// Pack Golang arrays of strings into Uint8 structures used by Postgres
//
func arrayStrToPg(a []string) (result []uint8) {
	s := "{"
	for _, value := range a {
		if s != "{" {
			s += ","
		}
		escaped := strings.Replace(value, "'", "''", -1)
		s += "'" + escaped + "'"
	}
	s += "}"

	return []byte(s)
}

// pgArrayIds is used to format a collection of integers into an ASCII rendering of
// a postgres SQL statement array
//
func pgArrayIds(ids []int64) (result string) {
	var outputBuffer bytes.Buffer
	for _, v := range ids {
		outputBuffer.WriteString(strconv.FormatInt(v, 10))
		outputBuffer.WriteString(",")
	}

	// Convert the Go style array into a slice then into a postgres array
	return fmt.Sprintf("{%s}", strings.Trim(outputBuffer.String(), ","))
}
