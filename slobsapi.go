package slobsapi

type RPCRequest struct {
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

func Request(id int, method, resource string, args ...interface{}) *RPCRequest {
	return &RPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params: &params{
			Resource: resource,
			Args:     args,
		},
	}
}

func RequestCompact(id int, method, resource string, args ...interface{}) *RPCRequest {
	return &RPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params: &params{
			Resource:    resource,
			Args:        args,
			CompactMode: true,
		},
	}
}
