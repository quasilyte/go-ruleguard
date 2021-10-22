package regression

type explicitIntType int

const (
	ZeroExplicit explicitIntType = iota // want `\Qgood, have explicit type`
	OneExplicit
	TwoExplicit
)

type implicitIntType int

// Matched due to the #295.
const typedIotaWithoutGroup int = iota // want `\Qgood, have explicit type`

const (
	ZeroImplicit = iota // want `\Qavoid use of iota without explicit type`
	OneImplicit
	TwoImplicit
)

type noIotaIntType int

// Matched due to the #295.
const iotaWithoutGroup = iota // want `\Qavoid use of iota without explicit type`

const (
	ZeroNoIota = 0
	OneNoIota  = 1
	TwoNoIota  = 2
)
