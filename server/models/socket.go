package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Conn   *websocket.Conn
	UserID int
}

type SocketManager struct {
	Sockets map[uint64]*websocket.Conn
	Mu      sync.RWMutex
}

type ConnectionType struct {
	Type string `json:"type"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	RoomID  string      `json:"roomId,omitempty"`
	GroupID uint64      `json:"groupId,omitempty"`
}

type BroadcastMessage struct {
	Data        interface{}
	TargetUsers map[uint64]bool // nil means broadcast to all
}

type EventRSVPMessage struct {
	Type     string `json:"type"`
	GroupID  int    `json:"groupId"`
	EventID  int    `json:"eventId"`
	Status   string `json:"status"`
	Going    int    `json:"going"`
	NotGoing int    `json:"notGoing"`
}
