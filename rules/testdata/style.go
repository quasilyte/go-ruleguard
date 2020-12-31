package target

func exprUnparen() {
	var f func(args ...interface{})

	f((1))      // want `\QexprUnparen: the parentheses around 1 are superfluous`
	f(1, ("x")) // want `\QexprUnparen: the parentheses around "x" are superfluous`
	f(1, ("y")) // want `\QexprUnparen: the parentheses around "y" are superfluous`
}

func emptyDecl() {
	var ()   // want `\QemptyDecl: empty var() block`
	const () // want `\QemptyDecl: empty const() block`
	type ()  // want `\QemptyDecl: empty type() block`
}
