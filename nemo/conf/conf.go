package conf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/lircstar/nemo/sys/log"
	"github.com/lircstar/nemo/sys/util"
)

type SYS struct {
	Version string `json:"version"`

	LittleEndian bool   `json:"little_endian"`
	LenStackBuf  int    `json:"len_stack_buf"`
	Monitor      string `json:"monitor"` // for bit operation. "1111" first bit : cpu second bit : mem third bit : block last bit : goroutine

	// log
	LogLevel string `json:"log_level"` // "debug" "info" "warn" "error" "fatal"
	LogFile  bool   `json:"log_file"`
	LogFlags string `json:"log_flags"`
}

type TCP struct {
	Addr            string `json:"addr"`
	InnerAddr       string `json:"inner_addr"`
	MaxConnNum      int    `json:"max_conn_num"`
	LenMsgLen       int    `json:"len_msg_len"`
	MinMsgLen       int    `json:"min_msg_len"`
	MaxMsgLen       int    `json:"max_msg_len"`
	TimeOut         int    `json:"time_out"`
	RoutineSafe     bool   `json:"routine_safe"`
	PendingWriteNum int    `json:"pending_write_num"`

	// Client
	Reconnect       bool          `json:"reconnect"`
	ConnectInterval time.Duration `json:"connect_interval"`
}

type UDP struct {
	Addr        string `json:"addr"`
	MaxConnNum  int    `json:"max_conn_num"`
	LenMsgLen   int    `json:"len_msg_len"`
	MaxMsgLen   int    `json:"max_msg_len"`
	MinMsgLen   int    `json:"min_msg_len"`
	TimeOut     int    `json:"time_out"`
	RoutineSafe bool   `json:"routine_safe"`

	// Client
	Reconnect       bool
	ConnectInterval time.Duration `json:"connect_interval"`
}

type WSS struct {
	Addr            string        `json:"addr"`
	CertFile        string        `json:"cert_file"`
	KeyFile         string        `json:"key_file"`
	MaxConnNum      int           `json:"max_conn_num"`
	MaxMsgLen       int           `json:"max_msg_len"`
	HTTPTimeout     time.Duration `json:"http_timeout"`
	PendingWriteNum int           `json:"pending_write_num"`

	// Client
	Reconnect bool
}

type Config struct {
	Sys SYS `json:"sys"`
	Tcp TCP `json:"tcp"`
	Udp UDP `json:"udp"`
	Wss WSS `json:"wss"`
}

var conf Config

func GetSYS() *SYS {
	return &conf.Sys
}

func GetTCP() *TCP {
	return &conf.Tcp
}

func GetUDP() *UDP {
	return &conf.Udp
}

func GetWSS() *WSS {
	return &conf.Wss
}

//var CallBack func(k string, v map[string]any)

func init() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("%v", err)
	}
	filename := filepath.Join(dir, "conf", util.GetProcessName()+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Setting default value
	conf.Sys.Version = "1.0.0"
	conf.Sys.LenStackBuf = 4096
	conf.Sys.Monitor = "0"      // for bit operation. "1111" first bit : cpu second bit : mem third bit : block last bit : goroutine
	conf.Sys.LogLevel = "debug" // "debug" "info" "warn" "error" "fatal"
	conf.Sys.LogFile = false
	conf.Sys.LittleEndian = false

	conf.Tcp.Addr = "127.0.0.1:6000"
	conf.Tcp.LenMsgLen = 2
	conf.Tcp.MinMsgLen = 1
	conf.Tcp.MaxMsgLen = 4096
	conf.Tcp.MaxConnNum = 65535
	conf.Tcp.TimeOut = 20
	conf.Tcp.RoutineSafe = true
	conf.Tcp.PendingWriteNum = 100

	conf.Tcp.Reconnect = false
	conf.Tcp.ConnectInterval = 3 * time.Second

	conf.Udp.Addr = "127.0.0.1:7000"
	conf.Udp.MaxConnNum = 65535
	conf.Udp.MinMsgLen = 1
	conf.Udp.MaxMsgLen = 4096
	conf.Udp.TimeOut = 10
	conf.Udp.RoutineSafe = true

	conf.Udp.Reconnect = false
	conf.Udp.ConnectInterval = 3 * time.Second

	conf.Wss.Addr = "127.0.0.1:6000"
	conf.Wss.MaxConnNum = 65535
	conf.Wss.MaxMsgLen = 4096
	conf.Wss.PendingWriteNum = 100
	conf.Wss.HTTPTimeout = 30 * time.Second

	conf.Wss.Reconnect = false

	err = json.Unmarshal(data, &conf)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
