package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"regexp"

	nex "github.com/NextendoNetwork/nextendo-nex"
)

// DL_TRACE=1 dumps every RMC request and response body (hex, plus any station URLs
// found inside) so new protocol behaviour can be read straight from the log.
var traceRMC = os.Getenv("DL_TRACE") == "1"

const traceMaxBytes = 768

var stationURLPattern = regexp.MustCompile(`(prudps?|udp):/[\x20-\x7e]+`)

func traceBody(b []byte) string {
	n := len(b)
	if n > traceMaxBytes {
		b = b[:traceMaxBytes]
	}
	s := fmt.Sprintf("len=%d %s", n, hex.EncodeToString(b))
	for _, u := range stationURLPattern.FindAll(b, -1) {
		s += "\n        url: " + string(u)
	}
	return s
}

// traced wraps a protocol handler so its responses are logged when tracing is on.
func traced(h nex.RMCHandler) nex.RMCHandler {
	if !traceRMC {
		return h
	}
	return func(c *nex.Connection, req *nex.RMCMessage) *nex.RMCMessage {
		resp := h(c, req)
		switch {
		case resp == nil:
			fmt.Printf("[trace] <- pid=%d proto=%#x method=%d call=%d (answered later)\n", c.PID, req.Protocol, req.Method, req.CallID)
		case resp.IsError:
			fmt.Printf("[trace] <- pid=%d proto=%#x method=%d call=%d ERROR %#x\n", c.PID, req.Protocol, req.Method, req.CallID, resp.Result)
		default:
			fmt.Printf("[trace] <- pid=%d proto=%#x method=%d call=%d %s\n", c.PID, req.Protocol, req.Method, req.CallID, traceBody(resp.Body))
		}
		return resp
	}
}

// traceRequest logs a request body (used from the endpoint's OnRMC hook).
func traceRequest(tag string, c *nex.Connection, req *nex.RMCMessage) {
	if traceRMC {
		fmt.Printf("[trace] -> %s pid=%d proto=%#x method=%d call=%d %s\n", tag, c.PID, req.Protocol, req.Method, req.CallID, traceBody(req.Body))
	}
}
