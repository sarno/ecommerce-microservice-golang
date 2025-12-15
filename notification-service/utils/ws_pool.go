package utils

import (
	"sync"

	"github.com/gorilla/websocket"
)

var (
	wsClient = make(map[int]*websocket.Conn)
	wsClientMutex = sync.RWMutex{}
)

func AddWebSocketClientConn(userId int, ws *websocket.Conn) {
	wsClientMutex.Lock()
	defer wsClientMutex.Unlock()
	wsClient[userId] = ws
}

func GetWebSocketClientConn(userId int) *websocket.Conn {
	wsClientMutex.RLock()
	defer wsClientMutex.RUnlock()
	return wsClient[userId]
}

func RemoveWebSocketClientConn(userId int) {
	wsClientMutex.Lock()
	defer wsClientMutex.Unlock()
	delete(wsClient, userId)
}