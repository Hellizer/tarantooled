package tarantooled

type requestWorker struct {
	requests map[reqID]*request
	c        *connection
}

func newRequestWorker(c *connection) *requestWorker {
	reqs := make(map[reqID]*request, 10)
	return &requestWorker{requests: reqs, c: c}
}

func (rw *requestWorker) in(r *request) {
	rw.requests[r.requestID] = r
	go rw.sender(r.requestID)
	go rw.reader()
}

func (rw *requestWorker) sender(id reqID) {
	rw.c.send(rw.requests[id])
}

func (rw *requestWorker) reader() {
	resp := rw.c.receive()
	//fmt.Println("got response")
	if resp != nil {
		rw.call(resp)
	}
}

func (rw *requestWorker) call(r *Response) {
	rw.remover(r)
}

func (rw *requestWorker) remover(r *Response) {
	a := rw.requests[r.requestID].callBack
	(*a)(r)
	delete(rw.requests, r.requestID)
}

func (rw *requestWorker) finish() {
	//дописать синхронизацию

	rw.c.c.Close()
}
