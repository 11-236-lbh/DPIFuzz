package scenario

import (
	m "masterthesis"
	"strings"
	"encoding/binary"
	"bytes"
)

const (
	NotAnsweringToVN = 1
	DidNotEchoVersion = 2
	LastTwoVersionsAreActuallySeal = 3
	Timeout = 4
)

const ForceVersionNegotiation = 0x1a2a3a4a

func RunVersionNegotiationScenario(host string, trace *m.Trace) string {
	trace.Scenario = "version_negotation"
	trace.ScenarioVersion = 1

	conn := m.NewConnection(host, strings.Split(host, ":")[0])
	conn.Version = ForceVersionNegotiation
	conn.SendClientInitialPacket()
	packet, err, buf := conn.ReadNextPacket()

	if err != nil {
		trace.ErrorCode = Timeout
	} else {
		if packet.Header().PacketType() != m.VersionNegotiation {
			trace.ErrorCode = NotAnsweringToVN
			trace.Results["received_packet_type"] = packet.Header().PacketType()
		} else {
			packet, _ := packet.(*m.VersionNegotationPacket)
			trace.Results["supported_versions"] = packet.SupportedVersions

			nVersions := len(packet.SupportedVersions)
			if nVersions > 1 {
				v1, v2 := packet.SupportedVersions[nVersions-2], packet.SupportedVersions[nVersions-1]
				hash := bytes.NewBuffer(make([]byte, 0, 8))
				binary.Write(hash, binary.BigEndian, v1)
				binary.Write(hash, binary.BigEndian, v2)

				_, err := m.NewCleartextCryptoState().Read.Open(nil, m.EncodeArgs(packet.Header().PacketNumber()), buf[m.LongHeaderSize:], buf[:m.LongHeaderSize])
				if err == nil {
					trace.ErrorCode = LastTwoVersionsAreActuallySeal
				}
			}

			echoed_version := packet.Header().(*m.LongHeader).Version
			if echoed_version != conn.Version {
				trace.ErrorCode = DidNotEchoVersion
				trace.Results["echoed_version"] = echoed_version
			}
		}
	}



	return strings.Split(conn.ConnectedIp().String(), ":")[0]
}