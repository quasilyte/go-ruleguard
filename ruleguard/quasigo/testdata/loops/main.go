package main

const one int = 1

func main() {
	// While-style loops.
	{
		i := 0
		for i < 5 {
			println(i)
			i++
		}
	}
	{
		i2 := 2
		for i2 < 2 {
			println(i2)
			i2++
		}
	}
}
