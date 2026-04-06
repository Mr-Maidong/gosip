package sip

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

var (
	bufferSize uint16 = 65535 - 20 - 8 // IPv4 max size - IPv4 Header size - UDP Header size
)

// RequestHandler RequestHandler
type RequestHandler func(req *Request, tx *Transaction)

// Server Server
type Server struct {
	udpaddr net.Addr
	tcpaddr net.Addr // 添加TCP地址
	conn    Connection

	// 添加TCP连接管理
	tcpConnections map[string]Connection // key: remote_addr, value: tcp_connection
	tcpConnMutex   *sync.RWMutex

	txs *transacionts

	hmu             *sync.RWMutex
	requestHandlers map[RequestMethod]RequestHandler

	port *Port
	host net.IP
}

// NewServer NewServer
func NewServer() *Server {
	activeTX = &transacionts{txs: map[string]*Transaction{}, rwm: &sync.RWMutex{}}
	srv := &Server{
		hmu:             &sync.RWMutex{},
		txs:             activeTX,
		requestHandlers: map[RequestMethod]RequestHandler{},
		tcpConnections:  make(map[string]Connection),
		tcpConnMutex:    &sync.RWMutex{},
	}
	return srv
}

// 添加TCP连接到管理器
func (s *Server) addTCPConnection(remoteAddr string, conn Connection) {
	s.tcpConnMutex.Lock()
	defer s.tcpConnMutex.Unlock()
	s.tcpConnections[remoteAddr] = conn
}

// 移除TCP连接
func (s *Server) removeTCPConnection(remoteAddr string) {
	s.tcpConnMutex.Lock()
	defer s.tcpConnMutex.Unlock()
	delete(s.tcpConnections, remoteAddr)
}

// 获取TCP连接
func (s *Server) getTCPConnection(remoteAddr string) (Connection, bool) {
	s.tcpConnMutex.RLock()
	defer s.tcpConnMutex.RUnlock()
	conn, ok := s.tcpConnections[remoteAddr]
	return conn, ok
}

// 根据目标地址和协议类型发送请求
func (s *Server) RequestWithProtocol(req *Request, protocol string) (*Transaction, error) {
	viaHop, ok := req.ViaHop()
	if !ok {
		return nil, fmt.Errorf("missing required 'Via' header")
	}
	viaHop.Host = s.host.String()
	viaHop.Port = s.port
	if viaHop.Params == nil {
		viaHop.Params = NewParams().Add("branch", String{Str: GenerateBranch()})
	}
	if !viaHop.Params.Has("rport") {
		viaHop.Params.Add("rport", nil)
	}

	var tx *Transaction
	if strings.ToLower(protocol) == "tcp" {
		// 使用TCP连接
		destAddr := req.Destination().String()
		if tcpConn, ok := s.getTCPConnection(destAddr); ok {
			tx = s.txs.newTX(getTXKey(req), tcpConn)
		} else {
			return nil, fmt.Errorf("TCP connection not found for %s", destAddr)
		}
	} else {
		// 使用UDP连接
		tx = s.mustTX(getTXKey(req))
	}

	return tx, tx.Request(req)
}

//	func (s *Server) newTX(key string) *Transaction {
//		return s.txs.newTX(key, s.conn)
//	}
func (s *Server) getTX(key string) *Transaction {
	return s.txs.getTX(key)
}
func (s *Server) mustTX(key string) *Transaction {
	tx := s.txs.getTX(key)
	if tx == nil {
		tx = s.txs.newTX(key, s.conn)
	}
	return tx
}

// ListenTCPServer ListenTCPServer
func (s *Server) ListenTCPServer(addr string) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logrus.Fatal("net.ResolveTCPAddr err", err, addr)
	}
	s.tcpaddr = tcpaddr

	listener, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		logrus.Fatal("net.ListenTCP err", err, addr)
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			logrus.Errorln("tcp.AcceptTCP err", err)
			continue
		}
		go s.handleTCPConnection(conn)
	}
}

func (s *Server) handleTCPConnection(conn *net.TCPConn) {
	defer conn.Close()

	// 注册TCP连接
	remoteAddr := conn.RemoteAddr().String()
	tcpConn := newTCPConnection(conn)
	s.addTCPConnection(remoteAddr, tcpConn)
	defer s.removeTCPConnection(remoteAddr)

	buf := make([]byte, bufferSize)
	parser := newParser()
	defer parser.stop()

	go s.handlerListenTCP(parser.out, tcpConn)

	// TCP流式数据缓冲区
	var streamBuffer []byte

	for {
		n, err := conn.Read(buf)
		if err != nil {
			logrus.Errorln("tcp.Read err", err)
			break
		}
		// 将读取的数据追加到缓冲区
		streamBuffer = append(streamBuffer, buf[:n]...)
		// 从缓冲区中提取完整的SIP消息
		for {
			completeMessage, remaining := s.extractCompleteSIPMessage(streamBuffer)
			if completeMessage == nil {
				// 没有完整消息，等待更多数据
				break
			}
			// 发送完整消息给解析器
			parser.in <- newPacket(completeMessage, conn.RemoteAddr())

			// 更新缓冲区，保留剩余数据
			streamBuffer = remaining
		}
	}
}

// 专门处理TCP消息的方法
func (s *Server) handlerListenTCP(msgs chan Message, tcpConn Connection) {
	var msg Message
	for {
		msg = <-msgs
		switch tmsg := msg.(type) {
		case *Request:
			req := tmsg
			req.SetDestination(s.tcpaddr) // 使用TCP地址
			s.handlerRequestTCP(req, tcpConn)
		case *Response:
			resp := tmsg
			resp.SetDestination(s.tcpaddr) // 使用TCP地址
			s.handlerResponseTCP(resp, tcpConn)
		default:
			logrus.Errorln("undefind msg type,", tmsg, msg.String())
		}
	}
}

func (s *Server) handlerRequestTCP(msg *Request, tcpConn Connection) {
	// 为TCP连接创建专门的Transaction
	tx := s.newTCPTX(getTXKey(msg), tcpConn)
	utils.LogSIPRequest(msg.Source().String(), msg.Method().String(), tx.key, msg.String())
	s.hmu.RLock()
	handler, ok := s.requestHandlers[msg.Method()]
	s.hmu.RUnlock()
	if !ok {
		logrus.Errorln("not found handler func,requestMethod:", msg.Method(), msg.String())
		go handlerMethodNotAllowed(msg, tx)
		return
	}

	go handler(msg, tx)
}

func (s *Server) handlerResponseTCP(msg *Response, tcpConn Connection) {
	tx := s.getTX(getTXKey(msg))
	if tx == nil {
		logrus.Infoln("not found tx. receive TCP response from:", msg.Source(), "message: \n", msg.String())
	} else {
		logrus.Traceln("receive TCP response from:", msg.Source(), "txKey:", tx.key, "message: \n", msg.String())
		tx.receiveResponse(msg)
	}
}

// 为TCP连接创建Transaction
func (s *Server) newTCPTX(key string, tcpConn Connection) *Transaction {
	return s.txs.newTX(key, tcpConn)
}

// 从流式数据中提取完整的SIP消息
func (s *Server) extractCompleteSIPMessage(buffer []byte) ([]byte, []byte) {
	if len(buffer) == 0 {
		return nil, buffer
	}

	// 查找消息头结束标志（\r\n\r\n）
	headerEndIndex := bytes.Index(buffer, []byte("\r\n\r\n"))
	if headerEndIndex == -1 {
		// 头部不完整，等待更多数据
		return nil, buffer
	}

	// 检查是否只是空的分隔符（无效消息）
	if headerEndIndex == 0 {
		// 跳过这个无效的分隔符，继续处理剩余数据
		remaining := buffer[4:] // 跳过 \r\n\r\n
		if len(remaining) == 0 {
			return nil, remaining
		}
		// 递归处理剩余数据
		return s.extractCompleteSIPMessage(remaining)
	}

	// 头部结束位置
	headerEnd := headerEndIndex + 4

	// 验证消息头是否包含有效的SIP起始行
	headerData := buffer[:headerEndIndex]
	if !s.isValidSIPHeader(headerData) {
		// 不是有效的SIP消息，跳过到下一个可能的消息
		remaining := buffer[headerEnd:]
		if len(remaining) == 0 {
			return nil, remaining
		}
		return s.extractCompleteSIPMessage(remaining)
	}

	// 解析Content-Length
	contentLength := s.parseContentLength(buffer[:headerEnd])

	// 计算完整消息的长度
	totalMessageLength := headerEnd + contentLength

	if len(buffer) < totalMessageLength {
		// 消息体不完整，等待更多数据
		return nil, buffer
	}

	// 提取完整消息
	completeMessage := make([]byte, totalMessageLength)
	copy(completeMessage, buffer[:totalMessageLength])

	// 返回完整消息和剩余数据
	remaining := buffer[totalMessageLength:]
	return completeMessage, remaining
}

// 验证是否是有效的SIP消息头
func (s *Server) isValidSIPHeader(headerData []byte) bool {
	if len(headerData) == 0 {
		return false
	}

	headerStr := string(headerData)
	lines := strings.Split(headerStr, "\r\n")

	if len(lines) == 0 {
		return false
	}

	// 检查第一行是否是有效的SIP请求或响应
	firstLine := strings.TrimSpace(lines[0])
	if len(firstLine) == 0 {
		return false
	}

	// 检查是否是SIP请求（METHOD URI SIP/2.0）
	if s.isValidSIPRequest(firstLine) {
		return true
	}

	// 检查是否是SIP响应（SIP/2.0 STATUS_CODE REASON）
	if s.isValidSIPResponse(firstLine) {
		return true
	}

	return false
}

// 检查是否是有效的SIP请求行
func (s *Server) isValidSIPRequest(line string) bool {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return false
	}

	// 检查方法名是否为有效的SIP方法
	method := strings.ToUpper(parts[0])
	validMethods := []string{"REGISTER", "INVITE", "ACK", "CANCEL", "BYE", "OPTIONS", "INFO", "PRACK", "UPDATE", "REFER", "NOTIFY", "SUBSCRIBE", "MESSAGE"}

	for _, validMethod := range validMethods {
		if method == validMethod {
			// 检查版本字段
			return strings.HasPrefix(strings.ToUpper(parts[2]), "SIP/")
		}
	}

	return false
}

// 检查是否是有效的SIP响应行
func (s *Server) isValidSIPResponse(line string) bool {
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return false
	}

	// 检查版本字段
	if !strings.HasPrefix(strings.ToUpper(parts[0]), "SIP/") {
		return false
	}

	// 检查状态码
	if _, err := strconv.Atoi(parts[1]); err != nil {
		return false
	}

	return true
}

// 解析Content-Length头
func (s *Server) parseContentLength(headerData []byte) int {
	headerStr := string(headerData)
	lines := strings.Split(headerStr, "\r\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			// 提取Content-Length值
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				lengthStr := strings.TrimSpace(parts[1])
				if length, err := strconv.Atoi(lengthStr); err == nil {
					return length
				}
			}
		}
		// 处理简写形式 "l:"
		if strings.HasPrefix(strings.ToLower(line), "l:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				lengthStr := strings.TrimSpace(parts[1])
				if length, err := strconv.Atoi(lengthStr); err == nil {
					return length
				}
			}
		}
	}

	// 如果没有找到Content-Length头，默认为0
	return 0
}

// ListenUDPServer ListenUDPServer
func (s *Server) ListenUDPServer(addr string) {
	udpaddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logrus.Fatal("net.ResolveUDPAddr err", err, addr)
	}
	s.port = NewPort(udpaddr.Port)
	s.host, err = utils.ResolveSelfIP()
	if err != nil {
		logrus.Fatal("net.ListenUDP resolveip err", err, addr)
	}
	udp, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		logrus.Fatal("net.ListenUDP err", err, addr)
	}
	s.conn = newUDPConnection(udp)
	var (
		raddr net.Addr
		num   int
	)
	buf := make([]byte, bufferSize)
	parser := newParser()
	defer parser.stop()
	go s.handlerListen(parser.out)
	for {
		num, raddr, err = s.conn.ReadFrom(buf)
		if err != nil {
			logrus.Errorln("udp.ReadFromUDP err", err)
			continue
		}
		parser.in <- newPacket(append([]byte{}, buf[:num]...), raddr)
	}
}

// RegistHandler RegistHandler
func (s *Server) RegistHandler(method RequestMethod, handler RequestHandler) {
	s.hmu.Lock()
	s.requestHandlers[method] = handler
	s.hmu.Unlock()
}
func (s *Server) handlerListen(msgs chan Message) {
	var msg Message
	for {
		msg = <-msgs
		switch tmsg := msg.(type) {
		case *Request:
			req := tmsg
			req.SetDestination(s.udpaddr)
			s.handlerRequest(req)
		case *Response:
			resp := tmsg
			resp.SetDestination(s.udpaddr)
			s.handlerResponse(resp)
		default:
			logrus.Errorln("undefind msg type,", tmsg, msg.String())
		}
	}
}
func (s *Server) handlerRequest(msg *Request) {
	tx := s.mustTX(getTXKey(msg))
	utils.LogSIPRequest(msg.Source().String(), msg.Method().String(), tx.key, msg.String())
	s.hmu.RLock()
	handler, ok := s.requestHandlers[msg.Method()]
	s.hmu.RUnlock()
	if !ok {
		logrus.Errorln("not found handler func,requestMethod:", msg.Method(), msg.String())
		go handlerMethodNotAllowed(msg, tx)
		return
	}

	go handler(msg, tx)
}

func (s *Server) handlerResponse(msg *Response) {
	tx := s.getTX(getTXKey(msg))
	if tx == nil {
		utils.LogSIPMessage(logrus.InfoLevel,
			fmt.Sprintf("not found tx. receive response from: %s", msg.Source()),
			msg.String())
	} else {
		utils.LogSIPResponse(msg.Source().String(), tx.key, msg.String())
		tx.receiveResponse(msg)
	}
}

// Request Request
func (s *Server) Request(req *Request) (*Transaction, error) {
	viaHop, ok := req.ViaHop()
	if !ok {
		return nil, fmt.Errorf("missing required 'Via' header")
	}
	viaHop.Host = s.host.String()
	viaHop.Port = s.port
	if viaHop.Params == nil {
		viaHop.Params = NewParams().Add("branch", String{Str: GenerateBranch()})
	}
	if !viaHop.Params.Has("rport") {
		viaHop.Params.Add("rport", nil)
	}

	tx := s.mustTX(getTXKey(req))
	return tx, tx.Request(req)
}

func handlerMethodNotAllowed(req *Request, tx *Transaction) {
	resp := NewResponseFromRequest("", req, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), []byte{})
	tx.Respond(resp)
}
