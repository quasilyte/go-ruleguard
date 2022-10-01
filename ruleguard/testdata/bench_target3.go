package main

import "sync"

type withMutex struct {
	x  int
	mu sync.Mutex
}

type embedsMutex struct {
	x int
	sync.Mutex
	y int
}

func _() {
	_ = withMutex{}
	_ = embedsMutex{}
	{
		x := withMutex{}
		y := embedsMutex{}
		_ = x
		_ = y
	}
	_ = &withMutex{}
	_ = &embedsMutex{}
	{
		x := &withMutex{}
		y := &embedsMutex{}
		_ = x
		_ = y
	}

	_ = withMutex{}
	_ = embedsMutex{}
	{
		x := withMutex{}
		y := embedsMutex{}
		_ = x
		_ = y
	}
	_ = &withMutex{}
	_ = &embedsMutex{}
	{
		x := &withMutex{}
		y := &embedsMutex{}
		_ = x
		_ = y
	}

	_ = withMutex{}
	_ = embedsMutex{}
	{
		x := withMutex{}
		y := embedsMutex{}
		_ = x
		_ = y
	}
	_ = &withMutex{}
	_ = &embedsMutex{}
	{
		x := &withMutex{}
		y := &embedsMutex{}
		_ = x
		_ = y
	}

	_ = withMutex{}
	_ = embedsMutex{}
	{
		x := withMutex{}
		y := embedsMutex{}
		_ = x
		_ = y
	}
	_ = &withMutex{}
	_ = &embedsMutex{}
	{
		x := &withMutex{}
		y := &embedsMutex{}
		_ = x
		_ = y
	}

	_ = withMutex{}
	_ = embedsMutex{}
	{
		x := withMutex{}
		y := embedsMutex{}
		_ = x
		_ = y
	}
	_ = &withMutex{}
	_ = &embedsMutex{}
	{
		x := &withMutex{}
		y := &embedsMutex{}
		_ = x
		_ = y
	}

	_ = withMutex{}
	_ = embedsMutex{}
	{
		x := withMutex{}
		y := embedsMutex{}
		_ = x
		_ = y
	}
	_ = &withMutex{}
	_ = &embedsMutex{}
	{
		x := &withMutex{}
		y := &embedsMutex{}
		_ = x
		_ = y
	}

	_ = withMutex{}
	_ = embedsMutex{}
	{
		x := withMutex{}
		y := embedsMutex{}
		_ = x
		_ = y
	}
	_ = &withMutex{}
	_ = &embedsMutex{}
	{
		x := &withMutex{}
		y := &embedsMutex{}
		_ = x
		_ = y
	}

	_ = withMutex{}
	_ = embedsMutex{}
	{
		x := withMutex{}
		y := embedsMutex{}
		_ = x
		_ = y
	}
	_ = &withMutex{}
	_ = &embedsMutex{}
	{
		x := &withMutex{}
		y := &embedsMutex{}
		_ = x
		_ = y
	}
}
