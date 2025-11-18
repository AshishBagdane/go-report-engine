package output

type OutputStrategy interface {
	Send(data []byte) error
}
