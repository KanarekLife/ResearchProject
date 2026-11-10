package claudecode

import (
	"encoding/json"
	"net/http"

	"mcbench/integrations/inference"
)

// The part of MCP's streamable HTTP transport Claude Code needs to list and
// call tools: JSON-RPC requests posted to one endpoint, answered as JSON.
const (
	jsonRPCVersion       = "2.0"
	methodInitialize     = "initialize"
	methodToolsList      = "tools/list"
	methodToolsCall      = "tools/call"
	methodPing           = "ping"
	errMethodNotFound    = -32601
	errInvalidParams     = -32602
	contentTypeJSON      = "application/json"
	defaultProtocolVersn = "2025-06-18"
)

// call is one tool call from Claude Code, answered on out.
type call struct {
	name string
	args any
	out  chan string
}

type mcpHandler struct {
	tools []inference.Tool
	calls chan<- *call
}

type rpcRequest struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *mcpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.ID == nil { // a notification
		w.WriteHeader(http.StatusAccepted)
		return
	}
	resp := map[string]any{"jsonrpc": jsonRPCVersion, "id": req.ID}
	result, rerr := h.handle(r, req)
	if rerr != nil {
		resp["error"] = rerr
	} else {
		resp["result"] = result
	}
	w.Header().Set("Content-Type", contentTypeJSON)
	json.NewEncoder(w).Encode(resp)
}

func (h *mcpHandler) handle(r *http.Request, req rpcRequest) (any, *rpcError) {
	switch req.Method {
	case methodInitialize:
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		json.Unmarshal(req.Params, &p)
		if p.ProtocolVersion == "" {
			p.ProtocolVersion = defaultProtocolVersn
		}
		return map[string]any{
			"protocolVersion": p.ProtocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]string{"name": mcpServer, "version": "1"},
		}, nil
	case methodToolsList:
		list := make([]map[string]any, len(h.tools))
		for i, t := range h.tools {
			list[i] = map[string]any{"name": t.Name, "description": t.Description, "inputSchema": t.Parameters}
		}
		return map[string]any{"tools": list}, nil
	case methodToolsCall:
		var p struct {
			Name      string `json:"name"`
			Arguments any    `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return nil, &rpcError{Code: errInvalidParams, Message: err.Error()}
		}
		if p.Arguments == nil {
			p.Arguments = map[string]any{}
		}
		c := &call{name: p.Name, args: p.Arguments, out: make(chan string, 1)}
		select {
		case h.calls <- c:
		case <-r.Context().Done():
			return nil, &rpcError{Code: errInvalidParams, Message: "cancelled"}
		}
		select {
		case out := <-c.out:
			return map[string]any{"content": []map[string]string{{"type": "text", "text": out}}}, nil
		case <-r.Context().Done():
			return nil, &rpcError{Code: errInvalidParams, Message: "cancelled"}
		}
	case methodPing:
		return map[string]any{}, nil
	}
	return nil, &rpcError{Code: errMethodNotFound, Message: "method not found: " + req.Method}
}
