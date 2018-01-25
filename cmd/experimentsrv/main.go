package main

// This file contains the implementation of an experiment service.  This service
// supports reflection in the same manner as does swagger style services.
// To access these facilities and to use a command line tool for testing the
// grpc_cli tool is used.  This tool can be installed using the instructions
// found at https://github.com/grpc/grpc/blob/master/doc/command_line_tool.md
//
// Testing this service can be done by starting the binary and then using commands
// such as:
//
// bins/opt/grpc_cli call localhost:3000 ai.sentient.EchoService.Echo "message: 'test'"
// connecting to localhost:3000
// message: "test"
// date_time {
// 	  seconds: 1513910233
// }
//
// Rpc succeeded with OK status
//
// Using the cli tool more detailed information can be uncovered, for example:
//
// bins/opt/grpc_cli ls localhost:3000 ai.sentient.EchoService Echo
// Echo
//
// bins/opt/grpc_cli ls localhost:3000 ai.sentient.EchoService/Echo --l
//   rpc Echo(ai.sentient.EchoRequest) returns (ai.sentient.EchoResponse) {}
//
// bins/opt/grpc_cli type localhost:3000 ai.sentient.EchoResponse
// message EchoResponse {
//  string message = 1[json_name = "message"];
//    .google.protobuf.Timestamp date_time = 2[json_name = "dateTime"];
// }
//
// bins/opt/grpc_cli type localhost:3000 google.protobuf.Timestamp
// message Timestamp {
//  int64 seconds = 1[json_name = "seconds"];
//    int32 nanos = 2[json_name = "nanos"];
// }
//
// ~/grpc/bins/opt/grpc_cli call localhost:30001 grpc.health.v1.Health/Check "service: 'experimentsrv'"
// connecting to localhost:30001
// status: SERVING
//
// Rpc succeeded with OK status
//
// To run this server in test mode you might use a command line such as:
//
// PGUSER=pl PGHOST=dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com PGDATABASE=platform cmd/experimentsrv/bin/experimentsrv
//
// To call the methods within the API you will need to first fetch a security token for the API and then use the grpc CLI command to exercise the API for example:
//
// export AUTH0_DOMAIN=sentientai.auth0.com
// export AUTH0_TOKEN=$(curl -s --request POST --url 'https://sentientai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj", "client_secret": "AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms", "audience": "http://api.sentient.ai/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' | jq -r '"\(.access_token)"')
//
// grpc_cli ls localhost:30001 ai.sentient.experiment.Service -l
// grpc_cli type localhost:30001 ai.sentient.experiment.GetRequest -l
// grpc_cli call localhost:30001 ai.sentient.experiment.Service.Get "id: 't'"  --metadata authorization:"Bearer $AUTH0_TOKEN"
//

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/SentientTechnologies/platform-services"
	"github.com/SentientTechnologies/platform-services/version"

	"github.com/karlmutch/envflag"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"github.com/SentientTechnologies/platform-services/experiment"
)

const serviceName = "experimentsrv"

var (
	logger = platform.NewLogger(serviceName)

	ipPort = flag.String("ip-port", "0.0.0.0:30001", "TCP/IP adapter IP and port to run this gRPC service on")
)

func usage() {
	fmt.Fprintln(os.Stderr, path.Base(os.Args[0]))
	fmt.Fprintln(os.Stderr, "usage: ", os.Args[0], "[arguments]      Sentient Technologies experiment service      ", version.GitHash, "    ", version.BuildTime)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Arguments:")
	fmt.Fprintln(os.Stderr, "")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Environment Variables:")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "options can be read for environment variables by changing dashes '-' to underscores")
	fmt.Fprintln(os.Stderr, "and using upper case letters.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "To control log levels the LOGXI env variables can be used, these are documented at https://github.com/mgutz/logxi")
}

func init() {
	flag.Usage = usage
}

// Go runtime entry point for production builds.  This function acts as an alias
// for the main.Main function.  This allows testing and code coverage features of
// go to invoke the logic within the server main without skipping important
// runtime initialization steps.  The coverage tools can then run this server as if it
// was a production binary.
//
// main will be called by the go runtime when the master is run in production mode
// avoiding this alias.
//
func main() {

	quitC := make(chan struct{})
	defer close(quitC)

	if !flag.Parsed() {
		// Use the go options parser to load command line options that have been set, and look
		// for these options inside the env variable table
		//
		envflag.Parse()
	}

	// This is the one check that does not get tested when the server is under test
	//
	if _, err := platform.NewExclusive(serviceName, quitC); err != nil {
		logger.Error(fmt.Sprintf("An instance of this process is already running %s", err.Error()))
		os.Exit(-1)
	}

	Main()
}

// Production style main that will invoke the server as a go routine to allow
// a very simple supervisor and a test wrapper to coexist in terms of our logic.
//
// When using test mode 'go test ...' this function will not, normally, be run and
// instead the EntryPoint function will be called avoiding some initialization
// logic that is not applicable when testing.  There is one exception to this
// and that is when the go unit test framework is linked to the master binary,
// using a TestRunMain build flag which allows a binary with coverage
// instrumentation to be compiled with only a single unit test which is,
// infact an alias to this main.
//
func Main() {

	fmt.Printf("%s built at %s, against commit id %s\n", os.Args[0], version.BuildTime, version.GitHash)

	if !flag.Parsed() {
		// Use the go options parser to load command line options that have been set, and look
		// for these options inside the env variable table
		//
		envflag.Parse()
	}

	doneC := make(chan struct{})
	quitC := make(chan struct{})

	if errs := EntryPoint(quitC, doneC); len(errs) != 0 {
		for _, err := range errs {
			logger.Error(err.Error())
		}
		os.Exit(-1)
	}

	// After starting the application message handling loops
	// wait until the system has shutdown
	//
	select {
	case <-quitC:
	}

	// Allow the quitC to be sent across the server for a short period of time before exiting
	time.Sleep(time.Second)
}

// EntryPoint enables both test and standard production infrastructure to
// invoke this server.
//
// quitC can be used by the invoking functions to stop the processing
// inside the server and exit from the EntryPoint function
//
// doneC is used by the EntryPoint function to indicate when it has terminated
// its processing
//
func EntryPoint(quitC chan struct{}, doneC chan struct{}) (errs []errors.Error) {

	defer func() {
		defer func() {
			recover()
		}()
		close(doneC)
	}()

	errs = []errors.Error{}

	// Supplying the context allows the client to pubsub to cancel the
	// blocking receive inside the run
	ctx, cancel := context.WithCancel(context.Background())

	// Setup a channel to allow a CTRL-C to terminate all processing.  When the CTRL-C
	// occurs we cancel the background msg pump processing pubsub mesages from
	// google, and this will also cause the main thread to unblock and return
	//
	stopC := make(chan os.Signal)
	go func() {
		defer cancel()

		select {
		case <-quitC:
			return
		case <-stopC:
			logger.Warn(errors.New("CTRL-C interrupted").With("stack", stack.Trace().TrimRuntime()).Error())
			close(quitC)
			return
		}
	}()

	signal.Notify(stopC, os.Interrupt, syscall.SIGTERM)

	// Initiate database processing.  Without the DB at least entering a retry state the
	// server will never run so immediately return if this cannot be started
	//
	dbMsgC, dbErrorC, err := experiment.StartDB(ctx.Done())
	if err != nil {
		errs = append(errs, err)
	} else {

		go func() {
			for {
				select {
				case msg := <-dbMsgC:
					logger.Info(msg)
				case errMsg := <-dbErrorC:
					if errMsg == nil {
						logger.Debug("error channel emptied")
						return
					}
					if errMsg.Fatal {
						logger.Fatal(fmt.Sprint(errMsg.Err))
						os.Exit(-5)
					} else {
						logger.Warn(fmt.Sprint(errMsg.Err))
					}
				}
			}
		}()
	}

	// Now check for any fatal errors before allowing the system to continue.  This allows
	// all errors that could have ocuured as a result of incorrect options to be flushed
	// out rather than having a frustrating single failure at a time loop for users
	// to fix things
	//
	if len(errs) != 0 {
		return errs
	}

	msg := fmt.Sprintf("git hash version %s", version.GitHash)
	logger.Info(msg)

	// Will start a go routine internally and send errors on the channel.
	// An error present on the channel implies that the REST server has
	// failed
	errC := runServer(ctx, serviceName, *ipPort)

	close(doneC)

	// Start a dummy service for now.  Normally this would be the production main processing loop,
	// or a collection of independently processing components
	func(ctx context.Context) {
		select {
		case <-ctx.Done():
		case err := <-errC:
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}(ctx)

	return nil
}
