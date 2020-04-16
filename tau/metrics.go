// Copyright 2015 The go-tau Authors
// This file is part of the go-tau library.
//
// The go-tau library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-tau library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-tau library. If not, see <http://www.gnu.org/licenses/>.

package tau

import (
	"github.com/Tau-Coin/taucoin-mobile-mining-go/metrics"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/p2p"
)

var (
	propTxnInPacketsMeter    = metrics.NewRegisteredMeter("tau/prop/txns/in/packets", nil)
	propTxnInTrafficMeter    = metrics.NewRegisteredMeter("tau/prop/txns/in/traffic", nil)
	propTxnOutPacketsMeter   = metrics.NewRegisteredMeter("tau/prop/txns/out/packets", nil)
	propTxnOutTrafficMeter   = metrics.NewRegisteredMeter("tau/prop/txns/out/traffic", nil)
	propHashInPacketsMeter   = metrics.NewRegisteredMeter("tau/prop/hashes/in/packets", nil)
	propHashInTrafficMeter   = metrics.NewRegisteredMeter("tau/prop/hashes/in/traffic", nil)
	propHashOutPacketsMeter  = metrics.NewRegisteredMeter("tau/prop/hashes/out/packets", nil)
	propHashOutTrafficMeter  = metrics.NewRegisteredMeter("tau/prop/hashes/out/traffic", nil)
	propBlockInPacketsMeter  = metrics.NewRegisteredMeter("tau/prop/blocks/in/packets", nil)
	propBlockInTrafficMeter  = metrics.NewRegisteredMeter("tau/prop/blocks/in/traffic", nil)
	propBlockOutPacketsMeter = metrics.NewRegisteredMeter("tau/prop/blocks/out/packets", nil)
	propBlockOutTrafficMeter = metrics.NewRegisteredMeter("tau/prop/blocks/out/traffic", nil)
	reqHeaderInPacketsMeter  = metrics.NewRegisteredMeter("tau/req/headers/in/packets", nil)
	reqHeaderInTrafficMeter  = metrics.NewRegisteredMeter("tau/req/headers/in/traffic", nil)
	reqHeaderOutPacketsMeter = metrics.NewRegisteredMeter("tau/req/headers/out/packets", nil)
	reqHeaderOutTrafficMeter = metrics.NewRegisteredMeter("tau/req/headers/out/traffic", nil)
	reqBodyInPacketsMeter    = metrics.NewRegisteredMeter("tau/req/bodies/in/packets", nil)
	reqBodyInTrafficMeter    = metrics.NewRegisteredMeter("tau/req/bodies/in/traffic", nil)
	reqBodyOutPacketsMeter   = metrics.NewRegisteredMeter("tau/req/bodies/out/packets", nil)
	reqBodyOutTrafficMeter   = metrics.NewRegisteredMeter("tau/req/bodies/out/traffic", nil)
	reqStateInPacketsMeter   = metrics.NewRegisteredMeter("tau/req/states/in/packets", nil)
	reqStateInTrafficMeter   = metrics.NewRegisteredMeter("tau/req/states/in/traffic", nil)
	reqStateOutPacketsMeter  = metrics.NewRegisteredMeter("tau/req/states/out/packets", nil)
	reqStateOutTrafficMeter  = metrics.NewRegisteredMeter("tau/req/states/out/traffic", nil)
	miscInPacketsMeter       = metrics.NewRegisteredMeter("tau/misc/in/packets", nil)
	miscInTrafficMeter       = metrics.NewRegisteredMeter("tau/misc/in/traffic", nil)
	miscOutPacketsMeter      = metrics.NewRegisteredMeter("tau/misc/out/packets", nil)
	miscOutTrafficMeter      = metrics.NewRegisteredMeter("tau/misc/out/traffic", nil)
)

// meteredMsgReadWriter is a wrapper around a p2p.MsgReadWriter, capable of
// accumulating the above defined metrics based on the data stream contents.
type meteredMsgReadWriter struct {
	p2p.MsgReadWriter     // Wrapped message stream to meter
	version           int // Protocol version to select correct meters
}

// newMeteredMsgWriter wraps a p2p MsgReadWriter with metering support. If the
// metrics system is disabled, this function returns the original object.
func newMeteredMsgWriter(rw p2p.MsgReadWriter) p2p.MsgReadWriter {
	if !metrics.Enabled {
		return rw
	}
	return &meteredMsgReadWriter{MsgReadWriter: rw}
}

// Init sets the protocol version used by the stream to know which meters to
// increment in case of overlapping message ids between protocol versions.
func (rw *meteredMsgReadWriter) Init(version int) {
	rw.version = version
}

func (rw *meteredMsgReadWriter) ReadMsg() (p2p.Msg, error) {
	// Read the message and short circuit in case of an error
	msg, err := rw.MsgReadWriter.ReadMsg()
	if err != nil {
		return msg, err
	}
	// Account for the data traffic
	packets, traffic := miscInPacketsMeter, miscInTrafficMeter
	switch {
	case msg.Code == BlockHeadersMsg:
		packets, traffic = reqHeaderInPacketsMeter, reqHeaderInTrafficMeter
	case msg.Code == BlockBodiesMsg:
		packets, traffic = reqBodyInPacketsMeter, reqBodyInTrafficMeter

	case rw.version >= tau63 && msg.Code == NodeDataMsg:
		packets, traffic = reqStateInPacketsMeter, reqStateInTrafficMeter

	case msg.Code == NewBlockHashesMsg:
		packets, traffic = propHashInPacketsMeter, propHashInTrafficMeter
	case msg.Code == NewBlockMsg:
		packets, traffic = propBlockInPacketsMeter, propBlockInTrafficMeter
	case msg.Code == TxMsg:
		packets, traffic = propTxnInPacketsMeter, propTxnInTrafficMeter
	}
	packets.Mark(1)
	traffic.Mark(int64(msg.Size))

	return msg, err
}

func (rw *meteredMsgReadWriter) WriteMsg(msg p2p.Msg) error {
	// Account for the data traffic
	packets, traffic := miscOutPacketsMeter, miscOutTrafficMeter
	switch {
	case msg.Code == BlockHeadersMsg:
		packets, traffic = reqHeaderOutPacketsMeter, reqHeaderOutTrafficMeter
	case msg.Code == BlockBodiesMsg:
		packets, traffic = reqBodyOutPacketsMeter, reqBodyOutTrafficMeter

	case rw.version >= tau63 && msg.Code == NodeDataMsg:
		packets, traffic = reqStateOutPacketsMeter, reqStateOutTrafficMeter

	case msg.Code == NewBlockHashesMsg:
		packets, traffic = propHashOutPacketsMeter, propHashOutTrafficMeter
	case msg.Code == NewBlockMsg:
		packets, traffic = propBlockOutPacketsMeter, propBlockOutTrafficMeter
	case msg.Code == TxMsg:
		packets, traffic = propTxnOutPacketsMeter, propTxnOutTrafficMeter
	}
	packets.Mark(1)
	traffic.Mark(int64(msg.Size))

	// Send the packet to the p2p layer
	return rw.MsgReadWriter.WriteMsg(msg)
}
