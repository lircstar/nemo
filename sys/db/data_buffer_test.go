package db

import (
	"testing"
	"time"
)

func TestPushLiveList(t *testing.T) {
	// Create a new DataBufferBox
	dbBox := NewDataBufferBox()

	// Create a new DataBuffer
	dataBuffer := &DataBuffer{
		dataID: 1,
		live:   time.Now(),
	}

	// Push the DataBuffer to the live list
	dbBox.pushLiveList(dataBuffer)

	// Check if the DataBuffer is in the live list
	if dbBox.liveListHead != dataBuffer {
		t.Errorf("Expected liveListHead to be %v, got %v", dataBuffer, dbBox.liveListHead)
	}

	if dbBox.liveListEnd != dataBuffer {
		t.Errorf("Expected liveListEnd to be %v, got %v", dataBuffer, dbBox.liveListEnd)
	}

	// Check if the DataBuffer is marked as in the list
	if !dataBuffer.list {
		t.Errorf("Expected dataBuffer.list to be true, got false")
	}
}
