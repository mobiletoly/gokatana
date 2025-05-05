package katapp

type ErrScope int

const (
	ErrUnknown ErrScope = iota
	ErrInternal
	ErrNotFound
	ErrInvalidInput
	ErrDuplicate
	ErrFailedExternalService
	ErrUnauthorized
)

type Err struct {
	Scope ErrScope
	Msg   string
}

func (e *Err) Error() string {
	return e.Msg
}

func NewErr(scope ErrScope, msg string) *Err {
	return &Err{Scope: scope, Msg: msg}
}
