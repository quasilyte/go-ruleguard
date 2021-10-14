package regression

import "testing"

func TestParallel(t *testing.T) { // want `\QParallel test`
	t.Parallel()
}

func TestNotParallel1(t *testing.T) { // want `\QNot a parallel test`
	t.Fatalf("This test should fail")
}

func TestNotParallel2(t *testing.T) {}
