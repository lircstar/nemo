package network


type ConnOption struct {

	//readTimeout int64
	//writeTimeout int64

	noDelay bool

}


//func (opt *ConnOption) SetConnTimeout(read, write int64) {
//
//	opt.readTimeout = read
//	opt.writeTimeout = write
//
//}
//
//func (opt *ConnOption) SetConnReadTimeout(read int64) {
//
//	opt.readTimeout = read
//
//}
//
//func (opt *ConnOption) GetConnReadTimeout() int64 {
//	return opt.readTimeout
//}
//
//func (opt *ConnOption) GetConnWriteTimeout() int64 {
//	return opt.writeTimeout
//}
//
//func (opt *ConnOption) SetConnWriteTimeout(write int64) {
//
//	opt.writeTimeout = write
//
//}

func (opt *ConnOption) SetNoDelay(delay bool) {

	opt.noDelay = delay

}