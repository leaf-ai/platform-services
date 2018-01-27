package main

// This file contains tests for the DB layer.  In order for the test to run
// the DB flags must be set appropriately

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"

	"github.com/golang/protobuf/proto"

	model "github.com/SentientTechnologies/platform-services/experiment"
	grpc "github.com/SentientTechnologies/platform-services/gen/experimentsrv"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"
)

func TestDBA(t *testing.T) {

	timeout := time.Duration(time.Minute)
	giveUp := time.Now().Add(timeout)
	msg := fmt.Sprintf("first attempt failed, retrying the DB connection for %v", timeout)

	err := errors.New("")
	for {
		select {
		case <-time.After(time.Second):
			if err = model.GetDBStatus(); err != nil {
				if time.Now().After(giveUp) {
					t.Error(errors.Wrap(err).With("stack", stack.Trace().TrimRuntime()))
					return
				}
				if len(msg) != 0 {
					t.Log(msg)
					msg = ""
				}
				continue
			}
			return
		}
	}
}

// diffExp is provided so that grpc clone operations which do not respect the
// representation of zero length arrays versus nil's can be worked around during
// testing
//
func diffExp(l *grpc.Experiment, r *grpc.Experiment) (diffs []string) {
	rc := proto.Clone(r).(*grpc.Experiment)
	if len(rc.InputLayers) == 0 {
		if l.InputLayers != nil {
			rc.InputLayers = map[uint32]*grpc.InputLayer{}
		}
	}
	if len(rc.OutputLayers) == 0 {
		if l.OutputLayers != nil {
			rc.OutputLayers = map[uint32]*grpc.OutputLayer{}
		}
	}

	return deep.Equal(l, rc)
}

func newTestExperiment() (out *grpc.Experiment) {
	return &grpc.Experiment{
		Uid:          "test-only-" + model.GetPseudoUUID(),
		Name:         "test-only-" + model.GetPseudoUUID(),
		Description:  "test-only-" + model.GetPseudoUUID(),
		InputLayers:  map[uint32]*grpc.InputLayer{},
		OutputLayers: map[uint32]*grpc.OutputLayer{},
	}
}

func TestDBExperimentSimple(t *testing.T) {

	if err := model.GetDBStatus(); err != nil {
		t.Error(errors.Wrap(err).With("stack", stack.Trace().TrimRuntime()))
		return
	}

	in := newTestExperiment()

	exp, err := model.InsertExperiment(in)
	if err != nil {
		t.Error(err)
		return
	}

	// To check equivalence between the supplied data and the apparently written
	// data copy the two fields we know changed into the original data and then
	// do the deep comparison
	in.Created = exp.Created
	if diff := diffExp(in, exp); len(diff) != 0 {
		t.Error(errors.New(strings.Join(diff, ", ")).With("stack", stack.Trace().TrimRuntime()))
		return
	}

	if err = model.DeactivateExperiment(in.Uid); err != nil {
		t.Error(err.With("uid", in.Uid))
		return
	}

	// Try reinserting and make sure it fails
	if _, err = model.InsertExperiment(in); err == nil {
		t.Error("failed tests due to reinsertion of a duplicate experiment working")
		return
	}
}
