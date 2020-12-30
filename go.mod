module github.com/quasilyte/go-ruleguard

go 1.15

require (
	github.com/google/go-cmp v0.5.2
	github.com/quasilyte/go-ruleguard/dsl v0.0.0-20201227095750-17ad251c41a9 // indirect
	golang.org/x/tools v0.0.0-20200812195022-5ae4c3c160a0
)

replace github.com/quasilyte/go-ruleguard/dsl v0.0.0-20201227095750-17ad251c41a9 => ./dsl
