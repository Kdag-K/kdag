package hashgraph
import "fmt"

// SelfParentError is used to differentiate errors that are normal when the
// hashgraph is being used corrently by multiple go-routines, from errors that
// should not be occuring event in a concurrent context.
type SelfParentError struct {
	msg    string
	normal bool
}

// NewSelfParentError creates a new SelfParentError
func NewSelfParentError(msg string, normal bool) SelfParentError {
	return SelfParentError{
		msg:    msg,
		normal: normal,
	}
}

// Error implements the Error interface
func (e SelfParentError) Error() string {
	return e.msg
}

// IsNormalSelfParentError checks that an error is of type SelfParentError and
// that it is normal.
func IsNormalSelfParentError(err error) bool {
	spErr, ok := err.(SelfParentError)
	return ok && spErr.normal
}

var usedCodes = map[string]*Error{}

func errorID(codespace string, code uint32) string {
	return fmt.Sprintf("%s:%d", codespace, code)
}

func getUsed(codespace string, code uint32) *Error {
	return usedCodes[errorID(codespace, code)]
}

func setUsed(err *Error) {
	usedCodes[errorID(err.codespace, err.code)] = err
}

