package db

import (
	"time"
)

const (
	DataLiveDefault         = 30 * time.Minute // 30m
	UpdateDelayDefault      = 5 * time.Minute  // 5m
	QuickUpdateDelayDefault = 3 * time.Second  // 3s
)

type DataBuffer struct {
	useFlag    bool
	dataID     uint
	live       time.Time
	update     time.Time
	changeFlag bool
	next       *DataBuffer
	last       *DataBuffer
	lock       bool
	list       bool

	// Placeholder for derived class implementation
	data any
}

func (db *DataBuffer) GetDataID() uint {
	return db.dataID
}

func (db *DataBuffer) SetData(data any) {
	db.data = data
}

func NewDataBuffer() *DataBuffer {
	return &DataBuffer{}
}

type DataBufferBox struct {
	dataBufferMap   map[uint]*DataBuffer
	liveListHead    *DataBuffer
	liveListEnd     *DataBuffer
	updateList      []*DataBuffer
	quickUpdateList []*DataBuffer
	allocateds      []*DataBuffer
	frees           []*DataBuffer
	liveTime        time.Duration
	updateTime      time.Duration
	quickUpdateTime time.Duration

	// Callback function to create a new DataBuffer
	CreateBufferCallBack func() *DataBuffer
	// Callback function to handle data
	OnDataUpdateCallBack func(data *DataBuffer)
}

func NewDataBufferBox() *DataBufferBox {
	return &DataBufferBox{
		dataBufferMap:   make(map[uint]*DataBuffer),
		liveTime:        DataLiveDefault,
		updateTime:      UpdateDelayDefault,
		quickUpdateTime: QuickUpdateDelayDefault,
	}
}

func (db *DataBufferBox) CreateData() *DataBuffer {
	return db.newBuffer()
}

func (db *DataBufferBox) DeleteData(data *DataBuffer) {
	db.removeFromLiveList(data)
	db.freeBuffer(data)
}

func (db *DataBufferBox) LockData(data *DataBuffer) {
	if data != nil {
		data.lock = true
	}
}

func (db *DataBufferBox) UnlockData(data *DataBuffer) {
	if data != nil {
		data.lock = false
		data.live = time.Now().Add(db.liveTime)
		if !data.list {
			data.next = nil
			data.last = nil
			db.pushLiveList(data)
		}
		if data.changeFlag {
			db.QuickUpdateData(data)
		}
	}
}

func (db *DataBufferBox) UpdateUnlockData(data *DataBuffer) {
	if data != nil {
		data.lock = false
		data.live = time.Now().Add(db.liveTime)
		if !data.list {
			data.next = nil
			data.last = nil
			db.pushLiveList(data)
		}
		db.QuickUpdateData(data)
	}
}

func (db *DataBufferBox) UpdateData(data *DataBuffer) {
	if data != nil {
		if !data.changeFlag {
			data.changeFlag = true
			data.update = time.Now().Add(db.updateTime)
			db.updateList = append(db.updateList, data)
		}
		if !data.lock {
			data.live = time.Now().Add(db.liveTime)
			db.pushLiveList(data)
		}
	}
}

func (db *DataBufferBox) QuickUpdateData(data *DataBuffer) {
	if data != nil {
		if !data.changeFlag {
			data.changeFlag = true
			data.update = time.Now().Add(db.quickUpdateTime)
			db.quickUpdateList = append(db.quickUpdateList, data)
		} else {
			data.update = time.Now().Add(db.quickUpdateTime)
			db.removeFromList(&db.updateList, data)
			db.removeFromList(&db.quickUpdateList, data)
			db.quickUpdateList = append(db.quickUpdateList, data)
		}
	}
}

func (db *DataBufferBox) GetData(dataID uint) *DataBuffer {
	if data, found := db.dataBufferMap[dataID]; found {
		return data
	}
	return nil
}

func (db *DataBufferBox) AddData(dataID uint, data *DataBuffer) bool {
	if pdata, found := db.dataBufferMap[dataID]; found {
		if pdata.lock {
			return false
		}
		db.DeleteData(pdata)
	}
	db.dataBufferMap[dataID] = data
	data.useFlag = true
	data.dataID = dataID
	data.live = time.Now()
	data.update = time.Now()
	data.changeFlag = false
	data.next = nil
	data.last = nil
	data.list = false
	data.lock = true
	return true
}

func (db *DataBufferBox) Loop() {
	now := time.Now()
	for len(db.quickUpdateList) > 0 {
		data := db.quickUpdateList[0]
		if data.update.Before(now) {
			data.changeFlag = false
			db.OnDataUpdate(data)
			db.quickUpdateList = db.quickUpdateList[1:]
		} else {
			break
		}
	}

	for len(db.updateList) > 0 {
		data := db.updateList[0]
		if data.update.Before(now) {
			data.changeFlag = false
			db.OnDataUpdate(data)
			db.updateList = db.updateList[1:]
		} else {
			break
		}
	}

	for db.liveListHead != nil {
		if db.liveListHead.live.Before(now) {
			data := db.liveListHead
			db.liveListHead = data.next
			if db.liveListHead != nil {
				db.liveListHead.last = nil
			}
			if !data.lock {
				delete(db.dataBufferMap, data.dataID)
				db.freeBuffer(data)
			}
			data.list = false
			data.next = nil
			data.last = nil
		} else {
			break
		}
	}
}

func (db *DataBufferBox) SetDataLiveTime(t uint) {
	db.liveTime = time.Duration(t) * time.Second
}

func (db *DataBufferBox) SetUpdateTime(t uint) {
	db.updateTime = time.Duration(t) * time.Second
}

func (db *DataBufferBox) pushLiveList(data *DataBuffer) {
	if !data.list {
		data.list = true
		if db.liveListHead == nil {
			data.next = nil
			db.liveListHead = data
			db.liveListEnd = data
		} else if data.next != nil && data.last != nil {
			data.last.next = data.next
			data.next.last = data.last
			data.next = nil
			data.last = db.liveListEnd
			db.liveListEnd.next = data
			db.liveListEnd = data
		} else if data.next != nil && data.last == nil {
			db.liveListHead = data.next
			data.next.last = nil
			data.next = nil
			data.last = db.liveListEnd
			db.liveListEnd.next = data
			db.liveListEnd = data
		} else if data.next == nil && data.last == nil {
			data.last = db.liveListEnd
			db.liveListEnd.next = data
			db.liveListEnd = data
		}
	}
}

func (db *DataBufferBox) removeFromLiveList(data *DataBuffer) {
	if data.changeFlag {
		data.changeFlag = false
		db.OnDataUpdate(data)
		db.removeFromList(&db.updateList, data)
		db.removeFromList(&db.quickUpdateList, data)
	}
	if data.list {
		data.list = false
		if data.next != nil && data.last != nil {
			data.last.next = data.next
			data.next.last = data.last
		} else if data.next != nil && data.last == nil {
			db.liveListHead = data.next
			data.next.last = nil
		} else if data.next == nil && data.last != nil {
			db.liveListEnd = data.last
			db.liveListEnd.next = nil
		} else if data.next == nil && data.last == nil && db.liveListHead == data {
			db.liveListHead = nil
		}
		data.next = nil
		data.last = nil
	}
	if data.useFlag {
		delete(db.dataBufferMap, data.dataID)
	}
}

func (db *DataBufferBox) newBuffer() *DataBuffer {
	var result *DataBuffer
	if len(db.frees) > 0 {
		result = db.frees[len(db.frees)-1]
		db.frees = db.frees[:len(db.frees)-1]
	} else {
		result = db.CreateBufferCallBack()
		db.allocateds = append(db.allocateds, result)
	}
	if result != nil {
		result.useFlag = false
		result.live = time.Now()
		result.update = time.Now()
		result.changeFlag = false
		result.next = nil
		result.last = nil
		result.list = false
		result.lock = false
	}
	return result
}

func (db *DataBufferBox) freeBuffer(data *DataBuffer) {
	db.frees = append(db.frees, data)
}

func (db *DataBufferBox) OnDataUpdate(data *DataBuffer) {
	// Placeholder for derived class implementation
}

func (db *DataBufferBox) removeFromList(list *[]*DataBuffer, data *DataBuffer) {
	for i, v := range *list {
		if v == data {
			*list = append((*list)[:i], (*list)[i+1:]...)
			break
		}
	}
}
