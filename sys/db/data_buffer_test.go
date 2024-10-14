package db

import (
	"testing"
	"time"
)

func TestNewDataBufferBox(t *testing.T) {
	dbBox := NewDataBufferBox()
	dbBox.SetCreateBufferCallBack(func() any {
		return "defaultMetaData"
	})
	dbBox.SetOnDataUpdateCallBack(func(data any) {
		// handle data update
	})

	if dbBox == nil {
		t.Errorf("Expected NewDataBufferBox to return a non-nil value")
	}
	if dbBox.liveTime != DataLiveDefault {
		t.Errorf("Expected liveTime to be %v, got %v", DataLiveDefault, dbBox.liveTime)
	}
	if dbBox.updateTime != UpdateDelayDefault {
		t.Errorf("Expected updateTime to be %v, got %v", UpdateDelayDefault, dbBox.updateTime)
	}
	if dbBox.quickUpdateTime != QuickUpdateDelayDefault {
		t.Errorf("Expected quickUpdateTime to be %v, got %v", QuickUpdateDelayDefault, dbBox.quickUpdateTime)
	}
}

func TestAddData(t *testing.T) {
	dbBox := NewDataBufferBox()
	dbBox.SetCreateBufferCallBack(func() any {
		return "defaultMetaData"
	})
	dbBox.SetOnDataUpdateCallBack(func(data any) {
		// handle data update
	})

	metaData := "testMetaData"
	dataID := uint64(1)

	success := dbBox.AddData(dataID, metaData)
	if !success {
		t.Errorf("Expected AddData to return true, got false")
	}

	data := dbBox.GetDataBuffer(dataID)
	if data == nil {
		t.Errorf("Expected data to be added to dataBufferMap")
	}
	if data.metaData != metaData {
		t.Errorf("Expected metaData to be %v, got %v", metaData, data.metaData)
	}
}

func TestDeleteData(t *testing.T) {
	dbBox := NewDataBufferBox()
	dbBox.SetCreateBufferCallBack(func() any {
		return "defaultMetaData"
	})
	dbBox.SetOnDataUpdateCallBack(func(data any) {
		// handle data update
	})

	metaData := "testMetaData"
	dataID := uint64(1)

	dbBox.AddData(dataID, metaData)
	dbBox.DeleteData(dataID)

	data := dbBox.GetDataBuffer(dataID)
	if data != nil {
		t.Errorf("Expected data to be deleted from dataBufferMap")
	}
}

func TestPushLiveList(t *testing.T) {
	dbBox := NewDataBufferBox()
	dbBox.SetCreateBufferCallBack(func() any {
		return "defaultMetaData"
	})
	dbBox.SetOnDataUpdateCallBack(func(data any) {
		// handle data update
	})

	dataBuffer := &DataBuffer{
		dataID: 1,
		live:   time.Now(),
	}

	dbBox.pushLiveList(dataBuffer)

	if dbBox.liveListHead != dataBuffer {
		t.Errorf("Expected liveListHead to be %v, got %v", dataBuffer, dbBox.liveListHead)
	}

	if dbBox.liveListEnd != dataBuffer {
		t.Errorf("Expected liveListEnd to be %v, got %v", dataBuffer, dbBox.liveListEnd)
	}

	if !dataBuffer.list {
		t.Errorf("Expected dataBuffer.list to be true, got false")
	}
}

func TestRemoveFromLiveList(t *testing.T) {
	dbBox := NewDataBufferBox()
	dbBox.SetCreateBufferCallBack(func() any {
		return "defaultMetaData"
	})
	dbBox.SetOnDataUpdateCallBack(func(data any) {
		// handle data update
	})

	dataBuffer := &DataBuffer{
		dataID: 1,
		live:   time.Now(),
	}

	dbBox.pushLiveList(dataBuffer)
	dbBox.removeFromLiveList(dataBuffer)

	if dbBox.liveListHead != nil {
		t.Errorf("Expected liveListHead to be nil, got %v", dbBox.liveListHead)
	}

	if dbBox.liveListEnd != nil {
		t.Errorf("Expected liveListEnd to be nil, got %v", dbBox.liveListEnd)
	}

	if dataBuffer.list {
		t.Errorf("Expected dataBuffer.list to be false, got true")
	}
}
