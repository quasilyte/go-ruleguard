// This package tests rules specified in `rules.go` of ruleguard itself.
package main

const () // want `empty const\(\) block`
var ()   // want `empty var\(\) block`
type ()  // want `empty type\(\) block`

func main() {}
