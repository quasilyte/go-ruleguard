package regression

import "testing"

func TestParallel(t *testing.T) { // want `Parallel test`
	t.Parallel()
}

func TestNotParallel1(t *testing.T) { // want `Not a parallel test`
	t.Fatalf("This test should fail")
}

func TestNotParallel2(t *testing.T) {}
