package websocket

import (
	"fmt"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/websocket"
)

// Server is a struct type containing every value needed for a websocket server
type Server struct {
	protocolHandler ProtocolHandler
	logger          log.Logger
	services        map[ProtoID]*ServiceDescription
}

// RegisterService adds the given ServiceDescription to the Server's services map
func (s *Server) RegisterService(sd *ServiceDescription) error {
	_, exist := s.services[sd.protocolName]
	if exist {
		return fmt.Errorf("Service Endpoint %s already exists", sd.protocolName)
	}

	s.services[sd.protocolName] = sd
	return nil
}

// GetService returns a ServiceDescription given its ProtoID or an error
func (s *Server) GetService(name ProtoID) (*ServiceDescription, error) {
	sd, exist := s.services[name]
	if !exist {
		return nil, fmt.Errorf("Service Description %s does not exists", name)
	}

	return sd, nil
}

// Serve starts the http transport for the websocket, listening on addr
func (s *Server) Serve(addr string) error {
	return http.ListenAndServe(addr, s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Log("err", err)
		return
	}

	s.logger.Log("conn", conn.RemoteAddr())
	go s.handleConnection(conn)
}

func (s *Server) handleConnection(conn *websocket.Conn) {
	defer conn.Close()
	for {
		messageType, request, err := conn.ReadMessage()
		if err != nil {
			return
		}

		srv, me, data, err := s.protocolHandler.Decode(request)
		if err != nil {
			err = conn.WriteMessage(messageType, []byte(err.Error()))
			if err != nil {
				s.logger.Log("err", err)
				return
			}
			continue
		}

		service, err := s.GetService(*srv)
		if err != nil {
			err = conn.WriteMessage(messageType, []byte(err.Error()))
			if err != nil {
				s.logger.Log("err", err)
				return
			}
			continue
		}

		handler, err := service.GetEndpointHandler(*me)
		if err != nil {
			err = conn.WriteMessage(messageType, []byte(err.Error()))
			if err != nil {
				s.logger.Log("err", err)
				return
			}
			continue
		}

		res, err := handler(data)
		if err != nil {
			err = conn.WriteMessage(messageType, []byte(err.Error()))
			if err != nil {
				s.logger.Log("err", err)
				return
			}
			continue
		}

		response, err := s.protocolHandler.Encode(srv, me, res)
		if err != nil {
			err = conn.WriteMessage(messageType, []byte(err.Error()))
			if err != nil {
				s.logger.Log("err", err)
				return
			}
			continue
		}

		err = conn.WriteMessage(messageType, response)
		if err != nil {
			s.logger.Log("err", err)
			return
		}
	}
}

// NewServer returns a pointer to a Server instance
func NewServer(
	ph ProtocolHandler,
	logger log.Logger,
) *Server {
	return &Server{
		protocolHandler: ph,
		logger:          logger,
		services:        make(map[ProtoID]*ServiceDescription),
	}
}