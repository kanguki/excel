checkShell:
	ls -a | grep examples

.phony: example
example: checkShell
	go run examples/main.go
