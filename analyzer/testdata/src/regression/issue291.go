package regression

type explicitIntType int

const (
	ZeroExplicit explicitIntType = iota // want `\Qavoid use of iota for constant types`
	OneExplicit
	TwoExplicit
)

type implicitIntType int

const (
	ZeroImplicit = iota // want `\Qavoid use of iota for constant types`
	OneImplicit
	TwoImplicit
)

type noIotaIntType int

const (
	ZeroNoIota = 0
	OneNoIota  = 1
	TwoNoIota  = 2
)
