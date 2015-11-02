package errs

type ErrVarNotSet struct {
	VariableName string
}

func (err *ErrVarNotSet) Error() string {
	return "environment variable not set: " + err.VariableName
}
