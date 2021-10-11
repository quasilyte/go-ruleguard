package golint

import "unsafe"

func shiftOverflow() {
	var ui8 uint8
	var i8 int8
	var ui16 uint16
	var ui32 uint32
	var ui64 uint64

	_ = i8 << 1
	_ = i8 << 4
	_ = i8 << 7
	_ = i8 << 8  // want `\Qi8 (8 bits) too small for shift of 8`
	_ = i8 << 10 // want `\Qi8 (8 bits) too small for shift of 10`

	_ = ui8 << 1
	_ = ui8 << 4
	_ = ui8 << 7
	_ = ui8 << 8  // want `\Qui8 (8 bits) too small for shift of 8`
	_ = ui8 << 10 // want `\Qui8 (8 bits) too small for shift of 10`

	_ = ui16 << 1
	_ = ui16 << 4
	_ = ui16 << 15
	_ = ui16 << 16 // want `\Qui16 (16 bits) too small for shift of 16`
	_ = ui16 << 17 // want `\Qui16 (16 bits) too small for shift of 17`

	_ = ui32 << 1
	_ = ui32 << 4
	_ = ui32 << 31
	_ = ui32 << 32 // want `\Qui32 (32 bits) too small for shift of 32`
	_ = ui32 << 33 // want `\Qui32 (32 bits) too small for shift of 33`

	_ = ui64 << 1
	_ = ui64 << 4
	_ = ui64 << 63
	_ = ui64 << 64       // want `\Qui64 (64 bits) too small for shift of 64`
	_ = ui64 << (63 + 1) // want `\Qui64 (64 bits) too small for shift of (63 + 1)`

	const oneIf64Bit = ^uint(0) >> 63
	const constTest2 = ^uint8(0) >> 63

	if false {
		// deadcode, no warnings
		_ = ui8 << 8
	} else {
		// not a deadcode
		_ = (ui8 + 1) << 9 // want `\Q(ui8 + 1) (8 bits) too small for shift of 9`
	}

	if true {
		// not a deadcode
		_ = (ui8 + 1) << 9 // want `\Q(ui8 + 1) (8 bits) too small for shift of 9`
	} else {
		// deadcode, no warnings
		_ = ui8 << 8
	}

	const TEN = 10

	if TEN > 30 {
		_ = ui8 << 8 // deadcode, no warnings
	} else if TEN > 20 {
		_ = ui8 << 8 // deadcode, no warnings
	} else if TEN > 10 {
		_ = ui8 << 8 // deadcode, no warnings
	} else {
		_ = (ui8 + 1) << 9 // want `\Q(ui8 + 1) (8 bits) too small for shift of 9`
	}

	if TEN > 30 {
		_ = ui8 << 8 // deadcode, no warnings
	} else if TEN > 20 {
		_ = ui8 << 8 // deadcode, no warnings
	} else if TEN >= 10 {
		_ = (ui8 + 1) << 9 // want `\Q(ui8 + 1) (8 bits) too small for shift of 9`
	} else {
		_ = (ui8 + 1) << 9 // deadcode, no warnings
	}

	if TEN == 10 {
		_ = (ui8 + 1) << 9 // want `\Q(ui8 + 1) (8 bits) too small for shift of 9`
	} else if TEN == 11 {
		_ = (ui8 + 1) << 9 // deadcode, no warnings
	} else {
		_ = (ui8 + 1) << 9 // deadcode, no warnings
	}

	if TEN == 15 {
		if TEN > 0 {
			_ = (ui8 + 1) << 9 // deadcode, no warnings
			if TEN == 10 {
				_ = (ui8 + 1) << 9 // deadcode, no warnings
			}
		} else {
			_ = (ui8 + 1) << 9 // deadcode, no warnings
		}
	}
}

func ShiftDeadCode2() {
	var i int
	const iBits = 8 * unsafe.Sizeof(i)

	if iBits <= 32 {
		if iBits == 16 {
			_ = i >> 8
		} else {
			_ = i >> 16
		}
	} else {
		_ = i >> 32
	}

	if iBits >= 64 {
		_ = i << 32
		if iBits == 128 {
			_ = i << 64
		}
	} else {
		_ = i << 16
	}

	if iBits == 64 {
		_ = i << 32
	}
}
