package channels

type MsgCh struct {
	Err     error
	Message string
	Level   string //"info", "warn", "error"
}
