package mode

import(
	_"fmt"
	_"io"
	_"net"
	_"bufio"
	_"strconv"
	"net"
	"io"
	"strconv"
)
/*=====================================================================================================*/
//      声明
/*=====================================================================================================*/

const (
	socks5Version = 5
	socks5AuthNone     = 0
	socks5AuthPassword = 2
	socks5AuthNoAccept = 0xff
	socks5AuthPasswordVer = 1
	socks5Connect = 1
	socks5IP4    = 1
	socks5Domain = 3
	socks5IP6    = 4
)
const (
	socks5Success                 = 0
	socks5GeneralFailure          = 1
	socks5ConnectNotAllowed       = 2
	socks5NetworkUnreachable      = 3
	socks5HostUnreachable         = 4
	socks5ConnectionRefused       = 5
	socks5TTLExpired              = 6
	socks5CommandNotSupported     = 7
	socks5AddressTypeNotSupported = 8
)


/*=====================================================================================================*/
//      server
/*=====================================================================================================*/
func serveSocks5Client(clientConn net.Conn, SS *SocketServer) {
	defer clientConn.Close()

	buff := make([]byte, 262)
	reply := []byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x22, 0x22}

	if _, err := io.ReadFull(clientConn, buff[:2]); err != nil {
		return
	}
	if buff[0] != socks5Version {
		reply[1] = socks5AuthNoAccept
		clientConn.Write(reply[:2])
		return
	}
	numMethod := buff[1]
	if _, err := io.ReadFull(clientConn, buff[:numMethod]); err != nil {
		return
	}
	reply[1] = socks5AuthNone
	if _, err := clientConn.Write(reply[:2]); err != nil {
		return
	}

	if _, err := io.ReadFull(clientConn, buff[:4]); err != nil {
		return
	}
	if buff[1] != socks5Connect {
		reply[1] = socks5CommandNotSupported
		clientConn.Write(reply)
		return
	}

	addressType := buff[3]
	addressLen := 0
	switch addressType {
	case socks5IP4:
		addressLen = net.IPv4len
	case socks5IP6:
		addressLen = net.IPv6len
	case socks5Domain:
		if _, err := io.ReadFull(clientConn, buff[:1]); err != nil {
			return
		}
		addressLen = int(buff[0])
	default:
		reply[1] = socks5AddressTypeNotSupported
		clientConn.Write(reply)
		return
	}
	host := make([]byte, addressLen)
	if _, err := io.ReadFull(clientConn, host); err != nil {
		return
	}
	if _, err := io.ReadFull(clientConn, buff[:2]); err != nil {
		return
	}
	hostStr := ""
	switch addressType {
	case socks5IP4, socks5IP6:
		ip := net.IP(host)
		hostStr = ip.String()
	case socks5Domain:
		hostStr = string(host)
	}
	port := uint16(buff[0])<<8 | uint16(buff[1])
	if port < 1 || port > 0xffff {
		reply[1] = socks5HostUnreachable
		clientConn.Write(reply)
		return
	}
	portStr := strconv.Itoa(int(port))

	hostStr = net.JoinHostPort(hostStr, portStr)
	dest, err := SS.Dial("tcp4", string(host))
	if err != nil {
		reply[1] = socks5ConnectionRefused
		clientConn.Write(reply)
		return
	}
	defer dest.Close()
	reply[1] = socks5Success
	if _, err := clientConn.Write(reply); err != nil {
		return
	}

	go func() {
		defer clientConn.Close()
		defer dest.Close()
		io.Copy(clientConn, dest)
	}()

	io.Copy(dest, clientConn)
}
