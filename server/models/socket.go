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