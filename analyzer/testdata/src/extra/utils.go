package extra

var sink interface{}

func sinkFunc(args ...interface{}) {}

func foo() int { return 19 }

func mightFail() error { return nil }

func newInt() *int { return new(int) }
