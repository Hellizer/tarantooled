package tarantooled

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net"
	"sync"
	"time"
)

//Connection struct
type connection struct {
	c         net.Conn
	lastError *ConnError
	writeMut  sync.Mutex
	readMut   sync.Mutex

	//writers
	//readers
}

func (conn *connection) encodePass(login string, saltB64 []byte) (auth *[]byte, err error) {
	salt := make([]byte, 32)
	_, err = base64.StdEncoding.Decode(salt, saltB64)
	if err != nil {
		return
	}
	step1 := sha1.Sum([]byte(login))
	step2 := sha1.Sum(step1[:])
	hash := sha1.New()
	hash.Write(salt[:20])
	hash.Write(step2[:])
	step3 := hash.Sum(nil)
	a := make([]byte, 20)
	for i := 0; i < 20; i++ {
		a[i] = step1[i] ^ step3[i]
	}
	return &a, nil
}

func (conn *connection) packAuth(body map[uint32]interface{}) *[]byte {
	var packet []byte
	packet = append(packet, []byte{
		0xce, 0, 0, 0, 0, // length
		0x82,    // 2 element map
		0, 0x07, // request code
		1, 0xce, 0, 0, 0, 0,
	}...)
	//packet = append(packet, encodeAuth(body.(map[uint32]interface{}))...)
	packet = append(packet, encode4bytes(0xdf, uint32(2))...)
	packet = append(packet, encode4bytes(0xce, 0x23)...)
	packet = append(packet, encodeString(body[0x23].(string))...)
	packet = append(packet, encode4bytes(0xce, 0x21)...)
	packet = append(packet, encode4bytes(0xdd, uint32(2))...)
	some := body[0x21].([]interface{})
	for i := 0; i < 2; i++ {
		packet = append(packet, encodeString(some[i].(string))...)
	}
	l := uint32(len(packet) - 5)
	(packet)[1] = byte(l >> 24)
	(packet)[2] = byte(l >> 16)
	(packet)[3] = byte(l >> 8)
	(packet)[4] = byte(l)
	return &packet
}

//Open establish connection
func (conn *connection) auth(login string, epass *[]byte) bool {

	body := map[uint32]interface{}{
		0x23: login,
		0x21: []interface{}{string("chap-sha1"), string(*epass)},
	}
	packed := conn.packAuth(body)
	conn.c.Write(*packed)
	resp := conn.receive()
	if resp != nil {
		return !resp.HasErrors
	}
	conn.c.Close()
	err := errors.New("not receved request")
	conn.lastError = &ConnError{goError: err, errBody: "Error with auth response"}
	return false
}

//Open establish connection
func (conn *connection) open(addr string, login, pass string) bool {
	if addr == "" || login == "" || pass == "" {
		err := errors.New("wrong connect parameters")
		conn.lastError = &ConnError{goError: err, errBody: "wrong connect parameters"}
		return false
	}

	c, err := net.DialTimeout("tcp", addr, time.Millisecond*3)
	if err != nil {
		conn.lastError = &ConnError{goError: err, errBody: "Not connected"}
		return false
	}
	conn.c = c
	buf := make([]byte, 128)
	_, err = c.Read(buf)
	if err != nil {
		conn.lastError = &ConnError{goError: err, errBody: "Cant read from socket"}
		return false
	}

	epass, err := conn.encodePass(pass, buf[64:108])
	if err != nil {
		conn.lastError = &ConnError{goError: err, errBody: "Error encode password"}
		return false
	}

	if !conn.auth(login, epass) {
		c.Close()
		err = errors.New("auth failed")
		conn.lastError = &ConnError{goError: err, errBody: "Wrong credentials"}
		return false
	}
	return true
}

//Close close connection
func (conn *connection) close() {
	conn.c.Close()
}

//LastError get last error
func (conn *connection) getError() *ConnError {
	return conn.lastError
}

//LastError get last error
func (conn *connection) send(r *request) bool {
	conn.writeMut.Lock()
	n, err := conn.c.Write(*r.packedBody)
	conn.writeMut.Unlock()
	if err != nil {
		conn.lastError = &ConnError{goError: err, errBody: "Error write request"}
		return false
	}
	if n != len(*r.packedBody) {
		err = errors.New("wrong writed size")
		conn.lastError = &ConnError{goError: err, errBody: "Error write request"}
		return false
	}
	return true
}

func (conn *connection) receive() *Response {
	resp := &Response{}

	//err := resp.fillResponse(conn.c)
	respSizeBuf := make([]byte, 5)
	conn.readMut.Lock()
	n, err := conn.c.Read(respSizeBuf)
	if err != nil || n == 0 || respSizeBuf[0] != 0xce {
		conn.lastError = &ConnError{goError: err, errBody: "Error read response"}
		return nil
	}
	var i uint32 = 0
	respLength := decodeUint64(&respSizeBuf, &i)
	resp.RawResponse = make([]byte, respLength)
	n, err = conn.c.Read(resp.RawResponse)
	conn.readMut.Unlock()
	if err != nil || n == 0 {
		conn.lastError = &ConnError{goError: err, errBody: "Error read response"}
		return nil
	}
	resp.unpackHeader()
	resp.unpackBody()
	return resp
}
