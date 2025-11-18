package output

import "fmt"

func NewConsoleOutput() OutputStrategy {
	return &ConsoleOutput{}
}

type ConsoleOutput struct{}

func (c *ConsoleOutput) Send(data []byte) error {
	fmt.Println(string(data))
	return nil
}
