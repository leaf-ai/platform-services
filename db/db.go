package db

// The db.go file contains a Go lang marshalling layer on top of the Postgres and
// Go SQL layers.  One of the responsibilities of the db.go module is also to
// enable Postgres arrays and the like to be marshalled into go data structures

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	sq "github.com/Masterminds/squirrel"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"
)

const ExpectedDBVersion int64 = 1

var (
	databaseHostPort    = flag.String("dbaddr", defaultDBHostPort(), "Host:Port pair for the database server")
	databaseUser        = flag.String("dbuser", defaultDBUser(), "User name for accessing the database")
	databaseName        = flag.String("dbname", defaultDBName(), "The 'Experiment' database name")
	databaseOptions     = flag.String("dboptions", "", "Postgres options for inclusion in the 'Experiment' DSN, for example -dboptions=options=\"-c statement_timeout=2min\"")
	databaseMaxConn     = flag.Int("dbmaxconn", 72, "Sets a limit for open connections the master will use to postgres")
	databaseMaxIdleConn = flag.Int("dbmaxidleconn", 8, "Sets a limit for open and idle connections the master will use to postgres")

	// dbDownErr is used within the db layer to record if the DB is down.  A simple circuit breaker is
	// used with the DB gofunc handler loop to retry connections on a regular basis.  The DB methods
	// will respect this flag and fail out any calls until the circuit breaker
	//
	dbDownErr = errors.New("DB has not yet been initialized")

	dBase *sqlx.DB
)

func defaultDBUser() string {
	name := os.Getenv("PGUSER")
	if name == "" {
		return "pl"
	}
	return name
}

func defaultDBName() string {
	name := os.Getenv("PGDATABASE")
	if name == "" {
		return "platform"
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

// Log is used to hold unstructured log entries that will be marshalled
//
type Log struct {
	Priority    syslog.Priority
	Nanoseconds int64
	Source      string
	Msg         string
}

// ExperimentData will hold data ready for being handled by the SQL layer for experiments.
//
type ExperimentData struct {
	ID          int64     `db:"id"`
	UID         int64     `db:"uid"`
	Created     time.Time `db:"created"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

func dbHasCorrectVersion(expect int64) (err error, ok bool, actual int64) {
	err = dBase.QueryRow("SELECT version FROM upgrades ORDER BY timestamp DESC LIMIT 1").Scan(&actual)

	return err, (expect == actual), actual
}

// CloseDB is used to close any existing database connections to the Experiment DB
//
func CloseDB() error {

	dbDownErr = errors.New("database has been closed intentionally").With("stack", stack.Trace().TrimRuntime())

	return dBase.Close()
}

// getPassFromPGPass matches the very first line that can be matched from the postgres password file as documented
// by https://www.postgresql.org/docs/9.5/static/libpq-pgpass.html
//
func getPassFromPGPass(passFile string, host string, port string, db string, user string) (pass string, err errors.Error) {

	file, errGo := os.Open(passFile)
	if errGo != nil {
		return pass, errors.Wrap(errGo).With("passfile", passFile).With("stack", stack.Trace().TrimRuntime())
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

	if errGo = scanner.Err(); errGo != nil {
		return pass, errors.Wrap(errGo, "pass file parsing failed").With("passfile", passFile).With("stack", stack.Trace().TrimRuntime())
	}

	// Password was not found inside the pgpas file
	return "", nil
}

type DBErrorMsg struct {
	Fatal bool
	Err   errors.Error
}

// StartDB is used to open and then attempt to test the connection to the
// database, which contains state information for components of the platform
// ecosystem
//
func StartDB(quitC <-chan struct{}) (msgC chan string, errorC chan *DBErrorMsg, err errors.Error) {

	msgC = make(chan string)
	errorC = make(chan *DBErrorMsg)

	// The following function does not create a Live database connection.  This is done later
	// during the normal server life cycle
	//
	if err := initDB(*databaseHostPort, *databaseUser); err != nil {
		return msgC, errorC, err
	}

	// On the first time the master is started the DB is left in a down state and is initialized
	// during the normal life cycle of the server.
	//
	dbDownErr = errors.New("database start not yet completed").With("stack", stack.Trace().TrimRuntime())

	// Start a runtime monitor for DB connections held open, a liveness check
	// and a circuit breaker
	//
	go func() {

		defer CloseDB()

		defer close(msgC)
		defer close(errorC)

		// Start by instantly trying to get the database up and going, this duration
		// will be reset once the main service loop has started below to a much
		// longer interval that is appropriate for casually checking if the DB
		// is live.
		//
		dbCheckTimer := time.Duration(time.Microsecond)

		for {
			select {
			case <-time.After(dbCheckTimer):
				dbRecovery := dbDownErr != nil
				if errGo := dBase.Ping(); errGo != nil {

					dbDownErr = errors.Wrap(errGo)

					msg := fmt.Sprint("database ", *databaseHostPort, " ", *databaseName, " is currently down")
					err := &DBErrorMsg{
						Fatal: false,
						Err:   errors.Wrap(errGo, msg).With("dbHostPort", *databaseHostPort).With("dbName", *databaseName).With("stack", stack.Trace().TrimRuntime()),
					}
					select {
					case errorC <- err:
					default:
					}
					dbCheckTimer = time.Duration(5 * time.Second)
					continue
				}

				if dbRecovery {
					select {
					case msgC <- fmt.Sprint("database startup / recovery has been performed ", *databaseHostPort, " name ", *databaseName):
					default:
					}
					// Check that the Database has the expected version
					err, ok, version := dbHasCorrectVersion(ExpectedDBVersion)
					if err == nil && !ok {
						msg := fmt.Sprint("db has the wrong version ", ExpectedDBVersion, " expected got version ", version, " instead")
						errMsg := &DBErrorMsg{
							Fatal: true,
							Err:   errors.New(msg).With("stack", stack.Trace().TrimRuntime()).With("dbExpectedVersion", ExpectedDBVersion).With("dbVersion", version),
						}
						select {
						case errorC <- errMsg:
						default:
						}
						// The receiver is responsible for stopping the server
						return
					}
					if err != nil {
						msg := "DB has no version marker, suspect or incorrect database schema"
						errMsg := &DBErrorMsg{
							Fatal: true,
							Err:   errors.Wrap(err, msg).With("stack", stack.Trace().TrimRuntime()).With("dbExpectedVersion", ExpectedDBVersion).With("dbVersion", version),
						}
						select {
						case errorC <- errMsg:
						default:
						}
						// The receiver is responsible for stopping the server
						return
					}
				}

				dbDownErr = nil
				msg := fmt.Sprint("database has ", dBase.Stats().OpenConnections, " connections ",
					" ", *databaseHostPort, " name ", *databaseName, " dbConnectionCount ", dBase.Stats().OpenConnections)
				select {
				case msgC <- msg:
				default:
				}

			case <-quitC:
				select {
				case msgC <- "database monitor stopped":
				default:
				}
				return
			}

			dbCheckTimer = time.Duration(15 * time.Second)
		}
	}()

	return msgC, errorC, nil
}

// Status is used to retrieve the main eco system status from a server.  This function
// will return an error value of nil if the DB is running for a useful error for why
// it might not be running
//
func GetDBStatus() (err errors.Error) {
	return dbDownErr
}

// initDB will setup the database configuration but does not actually create a live connection
// or validate the parameters supplied to it.
//
func initDB(url string, user string) (err errors.Error) {

	pgPassFile := os.Getenv("PGPASSFILE")
	if len(pgPassFile) == 0 {
		if pgPassHome := os.Getenv("HOME"); len(pgPassHome) != 0 {
			pgPassFile = filepath.Join(pgPassHome, ".pgpass")
		}
		if _, errGo := os.Stat(pgPassFile); os.IsNotExist(errGo) {
			pgPassFile = "~/.pgpass"
		}
		// If standard sensible locations dont work due to
		// shell issues in the salt startup try a hard coded but
		// well known location
		if _, errGo := os.Stat(pgPassFile); os.IsNotExist(errGo) {
			pgPassFile = "/opt/sentient/.pgpass"
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

	db, errGo := sql.Open("postgres", datasetName)
	if errGo != nil {
		return errors.Wrap(errGo).With("dbConnectString", safeDatasetName).With("stack", stack.Trace().TrimRuntime())
	}

	db.SetMaxOpenConns(*databaseMaxConn)
	db.SetMaxIdleConns(*databaseMaxIdleConn)

	dBase = sqlx.NewDb(db, "postgres")
	// The follow functor takes a gRPC/Thrift style name and converts it to a DB column style name
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
func DBShowAllTrace() (output []string, err errors.Error) {

	output = []string{}

	rows, errGo := dBase.Queryx("show all")

	if errGo != nil {
		return nil, errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())
	}
	defer rows.Close()

	type DBSetting struct {
		Name        string
		Setting     string
		Description string
	}

	for rows.Next() {
		aRow := &DBSetting{}
		if errGo = rows.StructScan(aRow); errGo != nil {
			return nil, errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())
		}
		msg := fmt.Sprintf("%s='%s' %s", aRow.Name, aRow.Setting, aRow.Description)
		output = append(output, fmt.Sprint("DB Setting ", msg))
	}
	return output, nil
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

			log.Fatal(fmt.Sprint("golang SQL failure return code found ", inErr.Error()))

			os.Exit(-5)
		}

		if driverErr, ok := inErr.(*pq.Error); ok {

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
				log.Fatal(fmt.Sprint("Postgres driver fatal error ", driverErr.Error()))

				os.Exit((classNumber + 10) * -1)
			}
			return nil
		}
	}

	return inErr
}

type Experiment struct {
	ID          int64
	Uid         string
	Created     time.Time
	Name        string
	Description string
}

func (data *Experiment) DeepCopy() (result *Experiment, err errors.Error) {

	mod := new(bytes.Buffer)

	enc := gob.NewEncoder(mod)
	dec := gob.NewDecoder(mod)
	result = &Experiment{}
	if errGo := enc.Encode(*data); errGo != nil {
		return nil, errors.Wrap(errGo).With(stack.Trace().TrimRuntime())
	}
	if errGo := dec.Decode(result); errGo != nil {
		return nil, errors.Wrap(errGo).With(stack.Trace().TrimRuntime())
	}
	return result, nil
}

// SelectExperiment is used to retrieve one or more experiments that have been registered
// with the service.
//
// If the id is specified, that is not 0, then the database specific identifier
// will be used to retrieve a single row.  If the id is not specified then the application
// unique identifier will be used to retrieve the experiment.
//
// The function can return both a nil, or an empty array for the first result,
// along with a nil for the error in the case the SQL query works but returns
// no data.
//
func SelectExperiment(id uint64, uid string) (result []Experiment, err errors.Error) {

	if dbDownErr != nil {
		return nil, dbDownErr
	}

	query := sq.Select("id,uid,created,name,description").
		From("experiments").
		OrderBy("uid DESC").
		PlaceholderFormat(sq.Dollar)

	if id > 0 {
		query = query.Where(sq.Eq{"id": id})
	} else {
		if len(uid) > 0 {
			query = query.Where(sq.Eq{"uid": uid})
		} else {
			return nil, errors.New("selecting an experiment requires either the DB id or the experiment unique id to be specified").With("stack", stack.Trace().TrimRuntime())
		}
	}

	sql, args, errGo := query.ToSql()
	if CheckIfFatal(errGo) != nil {
		return nil, errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime()).With("id", id).With("uid", uid)
	}

	rows, errGo := dBase.Queryx(sql, args...)
	if CheckIfFatal(errGo) != nil {
		return nil, errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime()).With("id", id).With("uid", uid)
	}
	defer rows.Close()

	result = []Experiment{}
	for rows.Next() {
		row := Experiment{}
		if errGo = rows.StructScan(&row); CheckIfFatal(errGo) != nil {
			return nil, errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime()).With("id", id).With("uid", uid)
		}

		result = append(result, row)
	}

	return result, nil
}

// InsertExperiment is used to insert a single dataset record into the
// postgres database. A unique ID will be generated by the insertion
// operation and this ID will be placed into the returned record so that the
// caller has a unique reference to the inserted data for use in either
// performing other database operations or making reference to the project
// from other records.
//
// The result returned this function does not share memory with the input data
// parameter and is a deep copy.  This is slight less efficent than returning the
// mutated input structure but safer in regards to potential side effects as a
// result of assumptions made by the caller.
//
//
func InsertExperiment(data *Experiment) (result *Experiment, err errors.Error) {

	if dbDownErr != nil {
		return nil, dbDownErr
	}

	if data.ID != 0 {
		return nil, errors.New("an insert operation must not have the id field set").With("stack", stack.Trace().TrimRuntime()).With("uid", data.Uid)
	}

	result, err = data.DeepCopy()

	query := sq.Insert("experiments").
		Columns("uid", "created", "name", "description").
		Values(data.Uid, data.Created, data.Name, data.Description).
		Suffix("RETURNING \"id\"").
		RunWith(dBase.DB).
		PlaceholderFormat(sq.Dollar)

	if errGo := query.QueryRow().Scan(&result.ID); CheckIfFatal(errGo) != nil {
		return nil, errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime()).With("uid", data.Uid)
	}

	return result, nil
}

// SaveLog is used to insert a logging record either originating from the master,
// or from a client into the logging table
//
func SaveLog(log *Log) (err errors.Error) {

	if dbDownErr != nil {
		return dbDownErr
	}

	message := log.Msg
	if len(message) > 127 {
		message = message[:127]
	}

	insert := sq.Insert("logs").
		Columns("priority", "timestamp", "source", "msg").
		Values(int32(log.Priority), time.Unix(0, log.Nanoseconds), log.Source, message).
		RunWith(dBase.DB).
		PlaceholderFormat(sq.Dollar)

	if _, errGo := insert.Exec(); CheckIfFatal(errGo) != nil {
		return errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())
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
