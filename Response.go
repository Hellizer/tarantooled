package tarantooled

import (
	"fmt"
	"net"
)

//Response .
type Response struct {
	RawResponse  []byte
	HasErrors    bool
	errorText    string
	requestID    reqID
	responseCode uint32
	i            uint32
	TempDecoded  []interface{}
}

func (r *Response) fillResponse(c net.Conn) error {
	respSizeBuf := make([]byte, 5)
	n, err := c.Read(respSizeBuf)
	if err != nil || n == 0 || respSizeBuf[0] != 0xce {
		return err
	}
	var i uint32 = 0
	respLength := decodeUint64(&respSizeBuf, &i)
	r.RawResponse = make([]byte, respLength)
	n, err = c.Read(r.RawResponse)
	if err != nil || n == 0 {
		return err
	}
	return nil
}

func (r *Response) unpackHeader() {
	l := r.RawResponse[0] & 0xf
	r.i = 1
	for ; l > 0; l-- {
		key := r.RawResponse[r.i]
		r.i++
		switch key {
		case 0:
			r.responseCode = uint32(decodeUint64(&r.RawResponse, &r.i))
			if r.responseCode != 0 {
				r.HasErrors = true
			}
		case 1:
			r.requestID = reqID(decodeUint64(&r.RawResponse, &r.i))
		default:
			_ = decodeUint64(&r.RawResponse, &r.i)
		}
	}
	//fmt.Println(r)
}

func (r *Response) unpackBody() {
	l := r.RawResponse[r.i] & 0xf
	r.i++
	for ; l > 0; l-- {
		key := r.RawResponse[r.i]
		r.i++

		switch key {
		case 0x30: //48
			r.decodeInterface()
		case 0x31: //49
			r.errorText = r.decodeString()
			fmt.Println(r.errorText)
		default:
			_ = decodeUint64(&r.RawResponse, &r.i)
		}
	}
}

func decodeUint64(r *[]byte, o *uint32) uint64 {
	if (*r)[*o] == 0xce {
		*o++ //надо делать проверку что не будет вылет по индексу
		var n uint32
		n = (uint32((*r)[*o]) << 24) | (uint32((*r)[*o+1]) << 16) | (uint32((*r)[*o+2]) << 8) | uint32((*r)[*o+3])
		*o += 4
		return uint64(n)
	}
	*o++ //надо делать проверку что не будет вылет по индексу
	var n uint64
	n = (uint64((*r)[*o]) << 56) | (uint64((*r)[*o+1]) << 48) | (uint64((*r)[*o+2]) << 40) | uint64((*r)[*o+3])<<32 | (uint64((*r)[*o+4]) << 24) |
		(uint64((*r)[*o+5]) << 16) | (uint64((*r)[*o+6]) << 8) | uint64((*r)[*o+7])
	*o += 8
	return n
}

func (r *Response) decodeString() string {
	b := r.RawResponse[r.i]
	r.i++
	var l byte
	if b >= 0xa0 && b <= 0xbf {
		l = b & 0x1f
	} else {
		panic("not iml")
	}

	s := r.RawResponse[r.i : r.i+uint32(l)]
	//s := a[*o:l +*o +1]
	//str := string(r.RawResponse[r.i : r.i+uint32(l)])
	str := string(s)
	r.i += uint32(l)
	//fmt.Printf("str: %s \n", str)
	return str
}

func (r *Response) decodeInterface() interface{} {
	b := r.RawResponse[r.i]
	r.i++
	//var l byte
	if b <= 0x7f || b >= 0xe0 { //число
		if int8(b) < 0 {
			// не то что нужно пока
		}
		return uint64(b) //0..127
	}
	if b >= 0x80 && b <= 0x8f { //мапа в заголовке такая
		// пока что опять не то что нужно
	}
	if b >= 0x90 && b <= 0x9f { //странный короткий массив
		l := int(b & 0xf)
		temp := make([]interface{}, l)
		for it := 0; it < l; it++ {
			temp = append(temp, r.decodeInterface())
		}
		return temp
	}
	if b >= 0xa0 && b <= 0xbf { //короткий стринг
		r.i--
		return r.decodeString()

	}
	//я уже знаю что там массив  поэтому перейду срау к нему тут пока все временное
	if b == 0xdd {
		//берем длинну
		n := (uint32(r.RawResponse[r.i]) << 24) | (uint32(r.RawResponse[r.i+1]) << 16) | (uint32(r.RawResponse[r.i+2]) << 8) | uint32(r.RawResponse[r.i+3])
		r.i += 4
		r.TempDecoded = make([]interface{}, n)
		r.TempDecoded = append(r.TempDecoded, r.decodeInterface())
	}
	return nil
}
