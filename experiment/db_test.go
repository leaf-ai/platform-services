package experiment

// This file contains tests for the DB layer.  In order for the test to run
// the DB flags must be set appropriately

import ()

func TestDBExperiment(t *testing.T) {
	if nil == GetDBStatus() {
		t.Error("database not currently active for testing")
	}

	t.Error("not yet implemented")
}
