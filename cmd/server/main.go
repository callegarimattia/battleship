// Package main is the entry point of the server.
package main

func main() {
	app := Application{}
	if err := app.Run(); err != nil {
		panic(err)
	}
}
