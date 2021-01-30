// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

// Do not log errors as unstructured fields.
// Before:
//   logger.Infof("Unable to create profile. Error: %v", err)
// After:
//   logger.Info("Unable to create profile", zap.Error(err))
func logWithUnstructuredError(m dsl.Matcher) {
	m.Import(`go.uber.org/zap`)

	m.Match(
		`$x.Info($*_, $z, $*_)`,
		`$x.Info($*_, $z.Error(), $*_)`,
	).
		Where(m["x"].Type.Is("*zap.SugaredLogger") && m["z"].Type.Implements(`error`)).
		Report("Errors must be logged as a structured field. Use zap.Error(err) or With(zap.Error(err))")
}
