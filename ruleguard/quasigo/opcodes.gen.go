// Code generated "gen_opcodes.go"; DO NOT EDIT.

package quasigo

//go:generate stringer -type=opcode -trimprefix=op
type opcode byte

const (
	opInvalid opcode = 0

	// Encoding: 0x01 (width=1)
	// Stack effect: (value) -> ()
	opPop opcode = 1

	// Encoding: 0x02 (width=1)
	// Stack effect: (x) -> (x x)
	opDup opcode = 2

	// Encoding: 0x03 index:u8 (width=2)
	// Stack effect: () -> (value)
	opPushParam opcode = 3

	// Encoding: 0x04 index:u8 (width=2)
	// Stack effect: () -> (value)
	opPushLocal opcode = 4

	// Encoding: 0x05 (width=1)
	// Stack effect: () -> (false)
	opPushFalse opcode = 5

	// Encoding: 0x06 (width=1)
	// Stack effect: () -> (true)
	opPushTrue opcode = 6

	// Encoding: 0x07 constid:u8 (width=2)
	// Stack effect: () -> (const)
	opPushConst opcode = 7

	// Encoding: 0x08 index:u8 (width=2)
	// Stack effect: (value) -> ()
	opSetLocal opcode = 8

	// Encoding: 0x09 index:u8 (width=2)
	// Stack effect: unchanged
	opIncLocal opcode = 9

	// Encoding: 0x0a index:u8 (width=2)
	// Stack effect: unchanged
	opDecLocal opcode = 10

	// Encoding: 0x0b (width=1)
	// Stack effect: (value) -> (value)
	opReturnTop opcode = 11

	// Encoding: 0x0c (width=1)
	// Stack effect: unchanged
	opReturnFalse opcode = 12

	// Encoding: 0x0d (width=1)
	// Stack effect: unchanged
	opReturnTrue opcode = 13

	// Encoding: 0x0e offset:i16 (width=3)
	// Stack effect: unchanged
	opJump opcode = 14

	// Encoding: 0x0f offset:i16 (width=3)
	// Stack effect: (cond:bool) -> ()
	opJumpFalse opcode = 15

	// Encoding: 0x10 offset:i16 (width=3)
	// Stack effect: (cond:bool) -> ()
	opJumpTrue opcode = 16

	// Encoding: 0x11 funcid:u16 (width=3)
	// Stack effect: (args...) -> (results...)
	opCallBuiltin opcode = 17

	// Encoding: 0x12 (width=1)
	// Stack effect: (value) -> (result:bool)
	opIsNil opcode = 18

	// Encoding: 0x13 (width=1)
	// Stack effect: (value) -> (result:bool)
	opIsNotNil opcode = 19

	// Encoding: 0x14 (width=1)
	// Stack effect: (value:bool) -> (result:bool)
	opNot opcode = 20

	// Encoding: 0x15 (width=1)
	// Stack effect: (x:int y:int) -> (result:bool)
	opEqInt opcode = 21

	// Encoding: 0x16 (width=1)
	// Stack effect: (x:int y:int) -> (result:bool)
	opNotEqInt opcode = 22

	// Encoding: 0x17 (width=1)
	// Stack effect: (x:int y:int) -> (result:bool)
	opGtInt opcode = 23

	// Encoding: 0x18 (width=1)
	// Stack effect: (x:int y:int) -> (result:bool)
	opGtEqInt opcode = 24

	// Encoding: 0x19 (width=1)
	// Stack effect: (x:int y:int) -> (result:bool)
	opLtInt opcode = 25

	// Encoding: 0x1a (width=1)
	// Stack effect: (x:int y:int) -> (result:bool)
	opLtEqInt opcode = 26

	// Encoding: 0x1b (width=1)
	// Stack effect: (x:string y:string) -> (result:bool)
	opEqString opcode = 27

	// Encoding: 0x1c (width=1)
	// Stack effect: (x:string y:string) -> (result:bool)
	opNotEqString opcode = 28

	// Encoding: 0x1d (width=1)
	// Stack effect: (x:string y:string) -> (result:string)
	opConcat opcode = 29

	// Encoding: 0x1e (width=1)
	// Stack effect: (x:int y:int) -> (result:int)
	opAdd opcode = 30

	// Encoding: 0x1f (width=1)
	// Stack effect: (x:int y:int) -> (result:int)
	opSub opcode = 31

	// Encoding: 0x20 (width=1)
	// Stack effect: (s:string from:int to:int) -> (result:string)
	opStringSlice opcode = 32

	// Encoding: 0x21 (width=1)
	// Stack effect: (s:string from:int) -> (result:string)
	opStringSliceFrom opcode = 33

	// Encoding: 0x22 (width=1)
	// Stack effect: (s:string to:int) -> (result:string)
	opStringSliceTo opcode = 34

	// Encoding: 0x23 (width=1)
	// Stack effect: (s:string) -> (result:int)
	opStringLen opcode = 35
)

type opcodeInfo struct {
	width int
}

var opcodeInfoTable = [256]opcodeInfo{
	opInvalid: {width: 1},

	opPop:             {width: 1},
	opDup:             {width: 1},
	opPushParam:       {width: 2},
	opPushLocal:       {width: 2},
	opPushFalse:       {width: 1},
	opPushTrue:        {width: 1},
	opPushConst:       {width: 2},
	opSetLocal:        {width: 2},
	opIncLocal:        {width: 2},
	opDecLocal:        {width: 2},
	opReturnTop:       {width: 1},
	opReturnFalse:     {width: 1},
	opReturnTrue:      {width: 1},
	opJump:            {width: 3},
	opJumpFalse:       {width: 3},
	opJumpTrue:        {width: 3},
	opCallBuiltin:     {width: 3},
	opIsNil:           {width: 1},
	opIsNotNil:        {width: 1},
	opNot:             {width: 1},
	opEqInt:           {width: 1},
	opNotEqInt:        {width: 1},
	opGtInt:           {width: 1},
	opGtEqInt:         {width: 1},
	opLtInt:           {width: 1},
	opLtEqInt:         {width: 1},
	opEqString:        {width: 1},
	opNotEqString:     {width: 1},
	opConcat:          {width: 1},
	opAdd:             {width: 1},
	opSub:             {width: 1},
	opStringSlice:     {width: 1},
	opStringSliceFrom: {width: 1},
	opStringSliceTo:   {width: 1},
	opStringLen:       {width: 1},
}
