package raw

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/lircstar/nemo/nemo/conf"
	"github.com/lircstar/nemo/nemo/network"
	"github.com/lircstar/nemo/sys/log"
)

// -------------------------
// | id | protobuf message |
// -------------------------
type Processor struct {
	littleEndian bool
	msgInfo      map[uint16]*MsgInfo
}

type MsgInfo struct {
	msgHandler network.MsgHandler
}

type Message struct {
	Id   uint16
	Data []byte
}

func NewProcessor() *Processor {
	p := new(Processor)
	p.littleEndian = conf.GetSYS().LittleEndian
	p.msgInfo = make(map[uint16]*MsgInfo)
	return p
}

// Register It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) Register(msg any) {
}

// SetHandler It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) SetHandler(msg any, msgHandler network.MsgHandler) {
	id, ok := msg.(uint16)
	if !ok || id == 0 {
		log.Fatalf("message %v not registered", msg)
	}
	if p.msgInfo[id] != nil {
		log.Fatalf("message %d is already registered", id)
	}
	p.msgInfo[id] = new(MsgInfo)
	p.msgInfo[id].msgHandler = msgHandler
}

// SetRawHandler It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) SetRawHandler(id uint16, msgRawHandler network.MsgHandler) {
}

// Route goroutine safe
func (p *Processor) Route(agent network.Agent, msg any, userData any) error {
	rawMsg := msg.(Message)
	msgInfo, ok := p.msgInfo[rawMsg.Id]
	if !ok {
		return fmt.Errorf("message %v not registered", msg)
	}

	if msgInfo.msgHandler != nil {
		msgInfo.msgHandler(agent, []any{rawMsg.Data, userData})
	}
	return nil
}

// Unmarshal goroutine safe
func (p *Processor) Unmarshal(data []byte) (any, error) {
	if len(data) < 2 {
		return nil, errors.New("protobuf data too short")
	}

	// id
	var id uint16
	if p.littleEndian {
		id = binary.LittleEndian.Uint16(data)
	} else {
		id = binary.BigEndian.Uint16(data)
	}

	// msg
	msgInfo := p.msgInfo[id]
	if msgInfo == nil {
		return nil, errors.New(fmt.Sprintf("message didn't define; %v", id))
	}
	return Message{id, data[2:]}, nil
}

// Marshal goroutine safe
func (p *Processor) Marshal(msg any) ([][]byte, error) {
	msgInfo := msg.(Message)

	id := make([]byte, 2)
	if p.littleEndian {
		binary.LittleEndian.PutUint16(id, msgInfo.Id)
	} else {
		binary.BigEndian.PutUint16(id, msgInfo.Id)
	}

	return [][]byte{id, msgInfo.Data}, nil
}

// Range goroutine safe
func (p *Processor) Range(f func(id uint16, name string)) {
	for id, _ := range p.msgInfo {
		f(id, "")
	}
}

func (p *Processor) GetMsgId(msg any) uint16 {
	return msg.(Message).Id
}
