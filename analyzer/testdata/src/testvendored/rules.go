// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules(m dsl.Matcher) {
	// Test a vendored dependency can be imported successfully and used in a Where statement.
	// Otherwise, the rules semantics are not important.

	// Import a dependency which happens to be vendored, i.e. it has been downloaded under the "vendor" directory
	// instead of being pulled from the Internet.
	// Using the .invalid name (RFC 2606) to ensure the unit test is not accidentally downloading files.
	m.Import(`github.invalid/globex/logging`)

	// This rule should match a function named 'Errorf' that has an argument that implements the 'error' interface.
	// In addition, we only want to match the Errorf() function implemented by the logging.Logger struct.
	// We don't want to match fmt.Errorf().
	// We also simulate the fact github.invalid/globex/logging is vendored, i.e. it is in the 'vendor' directory.
	// This means while analyzing the code, the AST type is '*testvendored/github.invalid/globex/logging/Logger',
	// and yet m["x"].Type.Is("*logging.Logger") should return true.
	m.Match(
		`$x.Errorf($fmt, $*_, $y, $*_)`,
		`$x.Errorf($fmt, $*_, $y.Error(), $*_)`,
	).
		Where(m["x"].Type.Is("*logging.Logger") && m["y"].Type.Implements("error")).
		Report(`Errors must be logged as a structured field`)

	// A test rule that matches any function named 'Errorf' such as logging.Logger.Errorf() or fmt.Errorf()
	m.Match(
		`$x.Errorf($fmt, $*_, $y, $*_)`,
		`$x.Errorf($fmt, $*_, $y.Error(), $*_)`,
	).
		Where(m["y"].Type.Implements("error")).
		Report(`nothing special, just testing the Errorf rule`)

}
