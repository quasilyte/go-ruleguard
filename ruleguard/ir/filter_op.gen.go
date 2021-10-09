// Code generated "gen_filter_op.go"; DO NOT EDIT.

package ir

const (
	FilterInvalidOp FilterOp = 0

	// !$Args[0]
	FilterNotOp FilterOp = 1

	// $Args[0] && $Args[1]
	FilterAndOp FilterOp = 2

	// $Args[0] || $Args[1]
	FilterOrOp FilterOp = 3

	// $Args[0] == $Args[1]
	FilterEqOp FilterOp = 4

	// $Args[0] != $Args[1]
	FilterNeqOp FilterOp = 5

	// $Args[0] > $Args[1]
	FilterGtOp FilterOp = 6

	// $Args[0] < $Args[1]
	FilterLtOp FilterOp = 7

	// $Args[0] >= $Args[1]
	FilterGtEqOp FilterOp = 8

	// $Args[0] <= $Args[1]
	FilterLtEqOp FilterOp = 9

	// m[$Value].Addressable
	// $Value type: string
	FilterVarAddressableOp FilterOp = 10

	// m[$Value].Pure
	// $Value type: string
	FilterVarPureOp FilterOp = 11

	// m[$Value].Const
	// $Value type: string
	FilterVarConstOp FilterOp = 12

	// m[$Value].Text
	// $Value type: string
	FilterVarTextOp FilterOp = 13

	// m[$Value].Line
	// $Value type: string
	FilterVarLineOp FilterOp = 14

	// m[$Value].Value.Int()
	// $Value type: string
	FilterVarValueIntOp FilterOp = 15

	// m[$Value].Type.Size
	// $Value type: string
	FilterVarTypeSizeOp FilterOp = 16

	// m[$Value].Filter($Args[0])
	// $Value type: string
	FilterVarFilterOp FilterOp = 17

	// m[$Value].Node.Is($Args[0])
	// $Value type: string
	FilterVarNodeIsOp FilterOp = 18

	// m[$Value].Object.Is($Args[0])
	// $Value type: string
	FilterVarObjectIsOp FilterOp = 19

	// m[$Value].Type.Is($Args[0])
	// $Value type: string
	FilterVarTypeIsOp FilterOp = 20

	// m[$Value].Type.Underlying().Is($Args[0])
	// $Value type: string
	FilterVarTypeUnderlyingIsOp FilterOp = 21

	// m[$Value].Type.ConvertibleTo($Args[0])
	// $Value type: string
	FilterVarTypeConvertibleToOp FilterOp = 22

	// m[$Value].Type.AssignableTo($Args[0])
	// $Value type: string
	FilterVarTypeAssignableToOp FilterOp = 23

	// m[$Value].Type.Implements($Args[0])
	// $Value type: string
	FilterVarTypeImplementsOp FilterOp = 24

	// m[$Value].Text.Matches($Args[0])
	// $Value type: string
	FilterVarTextMatchesOp FilterOp = 25

	// m.File.Imports($Value)
	// $Value type: string
	FilterFileImportsOp FilterOp = 26

	// m.File.PkgPath.Matches($Value)
	// $Value type: string
	FilterFilePkgPathMatchesOp FilterOp = 27

	// m.File.Name.Matches($Value)
	// $Value type: string
	FilterFileNameMatchesOp FilterOp = 28

	// $Value holds a function name
	// $Value type: string
	FilterFilterFuncRefOp FilterOp = 29

	// $Value holds a string constant
	// $Value type: string
	FilterStringOp FilterOp = 30

	// $Value holds an int64 constant
	// $Value type: int64
	FilterIntOp FilterOp = 31

	// m[`$$`].Node.Parent().Is($Args[0])
	FilterRootNodeParentIsOp FilterOp = 32
)

var filterOpNames = map[FilterOp]string{
	FilterInvalidOp:              `Invalid`,
	FilterNotOp:                  `Not`,
	FilterAndOp:                  `And`,
	FilterOrOp:                   `Or`,
	FilterEqOp:                   `Eq`,
	FilterNeqOp:                  `Neq`,
	FilterGtOp:                   `Gt`,
	FilterLtOp:                   `Lt`,
	FilterGtEqOp:                 `GtEq`,
	FilterLtEqOp:                 `LtEq`,
	FilterVarAddressableOp:       `VarAddressable`,
	FilterVarPureOp:              `VarPure`,
	FilterVarConstOp:             `VarConst`,
	FilterVarTextOp:              `VarText`,
	FilterVarLineOp:              `VarLine`,
	FilterVarValueIntOp:          `VarValueInt`,
	FilterVarTypeSizeOp:          `VarTypeSize`,
	FilterVarFilterOp:            `VarFilter`,
	FilterVarNodeIsOp:            `VarNodeIs`,
	FilterVarObjectIsOp:          `VarObjectIs`,
	FilterVarTypeIsOp:            `VarTypeIs`,
	FilterVarTypeUnderlyingIsOp:  `VarTypeUnderlyingIs`,
	FilterVarTypeConvertibleToOp: `VarTypeConvertibleTo`,
	FilterVarTypeAssignableToOp:  `VarTypeAssignableTo`,
	FilterVarTypeImplementsOp:    `VarTypeImplements`,
	FilterVarTextMatchesOp:       `VarTextMatches`,
	FilterFileImportsOp:          `FileImports`,
	FilterFilePkgPathMatchesOp:   `FilePkgPathMatches`,
	FilterFileNameMatchesOp:      `FileNameMatches`,
	FilterFilterFuncRefOp:        `FilterFuncRef`,
	FilterStringOp:               `String`,
	FilterIntOp:                  `Int`,
	FilterRootNodeParentIsOp:     `RootNodeParentIs`,
}
var filterOpFlags = map[FilterOp]uint64{
	FilterAndOp:    flagIsBinaryExpr,
	FilterOrOp:     flagIsBinaryExpr,
	FilterEqOp:     flagIsBinaryExpr,
	FilterNeqOp:    flagIsBinaryExpr,
	FilterGtOp:     flagIsBinaryExpr,
	FilterLtOp:     flagIsBinaryExpr,
	FilterGtEqOp:   flagIsBinaryExpr,
	FilterLtEqOp:   flagIsBinaryExpr,
	FilterStringOp: flagIsBasicLit,
	FilterIntOp:    flagIsBasicLit,
}
