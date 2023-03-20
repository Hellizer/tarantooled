package tarantooled

//SQLService interface
type SQLService interface {
	GetSQL(sql string, body CallBackSQL)
	Close()
}

//CallBackSQL callback for GetSQL function
type CallBackSQL func(resp *Response)

type service struct {
	rw *requestWorker
}

func (s *service) GetSQL(sql string, cback CallBackSQL) {
	req := &request{requestID: requestID.next(), callBack: &cback, body: sql, funcCode: 0x22, requestCode: 10}
	//fmt.Println(req)
	req.packBody()
	s.rw.in(req)
	//cback(req.response)
}

func (s *service) Close() {
	s.rw.finish()
}
