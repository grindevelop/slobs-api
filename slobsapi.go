package slobsapi

import (
	"errors"

	"github.com/buger/jsonparser"

	"github.com/gorilla/websocket"
)

type rpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type params struct {
	Resource    string        `json:"resource"`
	Args        []interface{} `json:"args,omitempty"`
	CompactMode bool          `json:"compactMode,omitempty"`
}

type RemoteConn struct {
	WebSocket *websocket.Conn
	response  chan []byte
	id        int
}

func request(id int, method, resource string, args []interface{}, compact bool) *rpcRequest {
	return &rpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params: &params{
			Resource:    resource,
			Args:        args,
			CompactMode: compact,
		},
	}
}

func (rc *RemoteConn) nextID() int {
	rc.id++
	if rc.id == 0 {
		rc.id++
	}
	return rc.id
}

func (rc *RemoteConn) ListenEvents(receive chan []byte) error {
	rc.response = make(chan []byte)
	for {
		_, message, err := rc.WebSocket.ReadMessage()
		if err != nil {
			return err
		}
		if s, _ := jsonparser.GetString(message, "result", "_type"); s == "EVENT" {
			result, _, _, _ := jsonparser.Get(message, "result")
			receive <- result
		} else {
			rc.response <- message
		}
	}
}

func (rc *RemoteConn) send(request *rpcRequest) ([]byte, error) {
	err := rc.WebSocket.WriteJSON(request)
	if err != nil {
		return nil, err
	}
	var response []byte
	if rc.response != nil {
		response = <-rc.response
	} else {
		err := rc.WebSocket.ReadJSON(&response)
		if err != nil {
			return nil, err
		}
	}
	if msg, err := jsonparser.GetString(response, "error", "message"); err == nil {
		return nil, errors.New(msg)
	}
	result, _, _, _ := jsonparser.Get(response, "result")
	return result, nil
}

func (rc *RemoteConn) Call(method, resource string, args ...interface{}) ([]byte, error) {
	return rc.send(request(rc.nextID(), method, resource, args, false))
}

func (rc *RemoteConn) CallCompact(method, resource string, args ...interface{}) ([]byte, error) {
	return rc.send(request(rc.nextID(), method, resource, args, true))
}

func (rc *RemoteConn) Notify(method, resource string, args ...interface{}) error {
	err := rc.WebSocket.WriteJSON(request(rc.nextID(), method, resource, args, false))
	return err
}
