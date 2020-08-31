package slobsapi

import (
	"errors"

	"github.com/valyala/fastjson"

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
	ws       *websocket.Conn
	response chan *fastjson.Value
	nextID   func() int
}

func idGenerator() func() int {
	var id int
	return func() int {
		if id < 1 {
			id = 1
		} else {
			id++
		}
		return id
	}
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

func Connect(urlStr string) (*RemoteConn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(urlStr, nil)
	return &RemoteConn{ws: conn, nextID: idGenerator()}, err
}

func (conn *RemoteConn) do(request *rpcRequest) (*fastjson.Value, error) {
	err := conn.ws.WriteJSON(request)
	if err != nil {
		return nil, err
	}
	var response *fastjson.Value
	if conn.response != nil {
		response = <-conn.response
	} else {
		_, message, err := conn.ws.ReadMessage()
		if err != nil {
			return nil, err
		}
		response, err = fastjson.ParseBytes(message)
		if err != nil {
			return nil, err
		}
	}
	if msg := response.GetStringBytes("error", "message"); msg != nil {
		return nil, errors.New(string(msg))
	}
	return response.Get("result"), nil
}

func (conn *RemoteConn) ListenEvents(receive chan *fastjson.Value) error {
	conn.response = make(chan *fastjson.Value)
	for {
		_, message, err := conn.ws.ReadMessage()
		if err != nil {
			return err
		}
		response, err := fastjson.ParseBytes(message)
		if err != nil {
			return err
		}
		if string(response.GetStringBytes("result", "_type")) == "EVENT" {
			receive <- response.Get("result")
		} else {
			conn.response <- response
		}
	}
}

func (conn *RemoteConn) Notify(method, resource string, args ...interface{}) error {
	err := conn.ws.WriteJSON(request(conn.nextID(), method, resource, args, false))
	return err
}

func (conn *RemoteConn) Call(method, resource string, args ...interface{}) (*fastjson.Value, error) {
	return conn.do(request(conn.nextID(), method, resource, args, false))
}

func (conn *RemoteConn) CallCompact(method, resource string, args ...interface{}) (*fastjson.Value, error) {
	return conn.do(request(conn.nextID(), method, resource, args, true))
}

func (conn *RemoteConn) Close() {
	conn.ws.Close()
}
