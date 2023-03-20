package tarantooled

import (
	"fmt"
	"sync"
)

type reqID uint32

var requestID reqID

func (r *reqID) next() reqID {
	requestIDLocker.Lock()
	result := requestID
	requestID++
	requestIDLocker.Unlock()
	return result
}
func (r *reqID) reset() {
	requestIDLocker.Lock()
	requestID = 1
	requestIDLocker.Unlock()
}

var requestIDLocker sync.Mutex

//Init cre
func Init(addr string, login, pass string) (s SQLService, err *ConnError) {

	requestID.reset()
	conn := connection{}
	if conn.open(addr, login, pass) {
		fmt.Println("connected")
		return &service{rw: newRequestWorker(&conn)}, nil
	}
	err = conn.getError()
	//fmt.Printf(conn.getError().String())
	return
}
