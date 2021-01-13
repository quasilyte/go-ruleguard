// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

func logrus(m dsl.Matcher) {
	m.Match(
		`$log.Error($*_, $err, $*_)`,
		`$log.Errorf($*_, $err, $*_)`,
	).
		Where(m["err"].Type.Is(`error`) && m["log"].Type.Implements(`github.com/sirupsen/logrus.FieldLogger`)).
		Report(`$log.WithError($err).Error(...)`)	
}

func loggerType(m dsl.Matcher) {
	m.Import("github.com/sirupsen/logrus")

	m.Match(`testLoggerType($x)`).
		Where(m["x"].Type.Is(`*logrus.Logger`)).
		Report("testLoggerType called with *logrus.Logger")
}
