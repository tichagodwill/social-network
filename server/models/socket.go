package models

import (
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Conn   *websocket.Conn
	UserID int
}

type SocketManager struct {
    SocketCounter atomic.Uint64
    Sockets       map[uint64]*websocket.Conn
    Mu            sync.Mutex
}