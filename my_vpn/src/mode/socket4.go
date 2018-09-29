package mode


import(
	_"fmt"
	_"io"
	_"net"
	_"bufio"
	_"strconv"
	"net"
	"bufio"
	"fmt"
	"io"
)

/*=====================================================================================================*/
//      声明
/*=====================================================================================================*/

var socks4Errors = []string{
	"",
	"request rejected or failed",
	"request rejected because SOCKS server cannot connect to identd on the client",
	"request rejected because the client program and identd report different user-ids",
}

const (
	socks4Version       = 4
	socks4Connect       = 1
	socks4Granted       = 90
	socks4Rejected      = 91
	socks4ConnectFailed = 92
	socks4UserIDInvalid = 93
)


/*=====================================================================================================*/
//      server
/*=====================================================================================================*/
func serveSOCKS4Client(clientConn net.Conn, SS *SocketServer) {
	defer clientConn.Close()

	reader := bufio.NewReader(clientConn)
	buff, err := reader.Peek(9)
	if err != nil {
		return
	}
	if buff[8] != 0 {
		if _, err = reader.ReadSlice(0); err != nil {
			return
		}
	}

	reply := make([]byte, 8)
	if buff[0] != socks4Version {
		reply[1] = socks4Rejected
		clientConn.Write(reply)
		return
	}
	if buff[1] != socks4Connect {
		reply[1] = socks4Rejected
		clientConn.Write(reply)
		return
	}
	port := uint16(buff[2])<<8 | uint16(buff[3])
	ip := buff[4:8]

	// 建立回源连接
	host := fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], port)
	dest, err := SS.Dial("tcp4", host)
	if err != nil {
		reply[1] = socks4ConnectFailed
		clientConn.Write(reply)
		return
	}
	defer dest.Close()

	reply[1] = socks4Granted
	if _, err = clientConn.Write(reply); err != nil {
		return
	}

	go func() {
		defer clientConn.Close()
		defer dest.Close()
		io.Copy(dest, clientConn)
	}()
	io.Copy(clientConn, dest)
}

