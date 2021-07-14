package shiny

import "golang.org/x/exp/shiny/screen"

func Main(f func(screen screen.Screen)) {
	f(&defaultScreen)
}
