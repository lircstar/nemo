package conf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"nemo/sys/log"
	"nemo/sys/util"
)

var (
	// system
	LenStackBuf        = 4096
	Monitor     string = "0" // for bit operation. "1111" first bit : cpu second bit : mem third bit : block last bit : goroutine

	// log

	LogLevel string = "debug" // "debug" "info" "warn" "error" "fatal"
	LogFile  bool   = false
	//LogFlag int

	// tcpserver
	TCPAddr   string = "127.0.0.1:6000"
	TCPInAddr string

	LittleEndian       bool          = true
	LenMsgLen          int           = 2
	TcpMinMsgLen       int           = 1
	TcpMaxMsgLen       int           = 4096
	TcpMaxConnNum      int           = 65536
	TcpTimeout         int           = 20 // socket read timeout seconds.
	TcpConnectInterval time.Duration = 3 * time.Second
	TcpRoutineSafe                   = true // message is routine safe.

	// tcpclient
	TcpReconnect bool = true

	// udpserver
	UdpMaxConnNum  int = 65536
	UdpTimeout     int = 10 // socket timeout after UDPTimeout seconds.
	UdpMinMsgLen   int = 1
	UdpMaxMsgLen   int = 4096
	UdpRoutineSafe     = true // message is routine safe.

	// udpclient
	UdpReconnect       bool          = true
	UdpConnectInterval time.Duration = 3 * time.Second

	//webserver
	WSAddr      string        = "127.0.0.1:8000"
	HTTPTimeout time.Duration = time.Second * 30
	CertFile    string
	KeyFile     string

	// cluster
	ListenAddr      string
	ConnAddrs       []string
	PendingWriteNum int = 100
)

var CallBack func(k string, v map[string]interface{})

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

	var jsMap map[string]interface{}
	err = json.Unmarshal(data, &jsMap)
	if err != nil {
		log.Fatalf("%v", err)
	}

	for k, v := range jsMap {
		if k == "sys" {
			sys := v.(map[string]interface{})
			for k1, v1 := range sys {
				if k1 == "monitor" {
					Monitor = v1.(string)
				} else if k1 == "log_level" {
					LogLevel = v1.(string)
				} else if k1 == "log_file" {
					LogFile = v1.(bool)
				} else {
					log.Warnf("sys define invalidate json key : %v", k1)
				}
			}
		}

		if k == "tcp" {
			tcp := v.(map[string]interface{})
			for k1, v1 := range tcp {
				if k1 == "addr" {
					TCPAddr = v1.(string)
				} else if k1 == "inner_addr" {
					TCPInAddr = v1.(string)
				} else if k1 == "max_conn_num" {
					TcpMaxConnNum = int(v1.(float64))
				} else if k1 == "max_msg_len" {
					TcpMaxMsgLen = int(v1.(float64))
				} else if k1 == "time_out" {
					TcpTimeout = int(v1.(float64))
				} else if k1 == "little_endian" {
					LittleEndian = v1.(bool)
				} else if k1 == "reconnect" {
					TcpReconnect = v1.(bool)
				} else if k1 == "reconnect" {
					TcpConnectInterval = time.Duration(v1.(float64)) * time.Second
				} else if k1 == "routine_safe" {
					TcpRoutineSafe = v1.(bool)
				} else {
					log.Warnf("tcp define invalidate json key : %v", k1)
				}
			}
		}

		if k == "udp" {
			udp := v.(map[string]interface{})
			for k1, v1 := range udp {
				if k1 == "time_out" {
					UdpTimeout = int(v1.(float64))
				} else if k1 == "max_msg_len" {
					UdpMaxMsgLen = int(v1.(float64))
				} else if k1 == "reconnect" {
					UdpReconnect = v1.(bool)
				} else if k1 == "little_endian" {
					LittleEndian = v1.(bool)
				} else if k1 == "connect_interval" {
					UdpConnectInterval = time.Duration(v1.(float64)) * time.Second
				} else if k1 == "udp_max_conn" {
					UdpMaxConnNum = int(v1.(float64))
				} else if k1 == "routine_safe" {
					UdpRoutineSafe = v1.(bool)
				} else {
					log.Warnf("udp define invalidate json key : %v", k1)
				}
			}
		}

		if k == "web" {
			udp := v.(map[string]interface{})
			for k1, v1 := range udp {
				if k1 == "addr" {
					WSAddr = v1.(string)
				} else if k1 == "little_endian" {
					LittleEndian = v1.(bool)
				} else if k1 == "cert_file" {
					CertFile = v1.(string)
				} else if k1 == "key_file" {
					KeyFile = v1.(string)
				} else if k1 == "timeout" {
					HTTPTimeout = time.Second * time.Duration(int(v1.(float64)))
				} else {
					log.Warnf("web define invalidate json key : %v", k1)
				}
			}
		}

		if CallBack != nil {
			CallBack(k, v.(map[string]interface{}))
		}
	}
}
