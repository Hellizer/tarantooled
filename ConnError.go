package tarantooled

import "fmt"

//ConnError connection errors
type ConnError struct {
	goError error
	errCode uint32
	errBody string
}

func (ce ConnError) String() string {
	return fmt.Sprintf("Error body: %s, Base error: %s \n", ce.errBody, ce.goError.Error())
}
