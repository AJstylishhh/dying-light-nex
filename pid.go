package main

import (
	"math"
	"strconv"

	nex "github.com/NextendoNetwork/nextendo-nex"
)

// withUnsignedPIDs fixes the PID written into station URLs by core v0.2.1, which formats it
// with int(): a real Switch's 64-bit account id (>= 2^63) comes out negative, so peers get a
// host station that names nobody. Rewrites the stored stations and Register's public URL.
func withUnsignedPIDs(h nex.RMCHandler) nex.RMCHandler {
	return func(c *nex.Connection, req *nex.RMCMessage) *nex.RMCMessage {
		resp := h(c, req)
		if c.PID <= math.MaxInt64 {
			return resp
		}
		pid := strconv.FormatUint(c.PID, 10)

		stations := c.Stations()
		fixed := make([]*nex.StationURL, 0, len(stations))
		for _, st := range stations {
			u := st.Copy()
			u.Set("PID", pid)
			fixed = append(fixed, u)
		}
		c.SetStations(fixed)

		if resp != nil && !resp.IsError && (req.Method == nex.MethodRegister || req.Method == nex.MethodRegisterEx) {
			in := nex.NewStreamIn(resp.Body, c.Settings)
			result, cid, url := in.U32(), in.U32(), in.String()
			if in.Err() == nil {
				u := nex.ParseStationURL(url)
				u.Set("PID", pid)
				out := nex.NewStreamOut(c.Settings)
				out.U32(result)
				out.U32(cid)
				out.String(u.String())
				resp.Body = out.Bytes()
			}
		}
		return resp
	}
}
