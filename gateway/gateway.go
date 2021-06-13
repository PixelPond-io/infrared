package gateway

import (
	"errors"

	"github.com/haveachin/infrared/connection"
	"github.com/haveachin/infrared/protocol/handshaking"
)

var (
	ErrCantGetHSPacket   = errors.New("cant get handshake packet from caller")
	ErrNoServerFound     = errors.New("no server was found for this request")
	ErrNotValidHandshake = errors.New("this connection didnt provide a valid handshake")
)

func NewBasicGatewayWithStore(store ServerStore, ch <-chan connection.HandshakeConn) BasicGateway {
	return BasicGateway{store: store, inCh: ch}
}

type ServerData struct {
	ConnCh chan<- connection.HandshakeConn
}

type BasicGateway struct {
	store ServerStore
	inCh  <-chan connection.HandshakeConn
}

func (g *BasicGateway) Start() error {
	for {
		conn := <-g.inCh
		g.HandleConn(conn)
	}
}

func (g *BasicGateway) HandleConn(conn connection.HandshakeConn) error {
	pk, err := conn.ReadPacket()
	if err != nil {
		return ErrCantGetHSPacket
	}

	hs, err := handshaking.UnmarshalServerBoundHandshake(pk)
	if err != nil {
		return err
	}
	addr := string(hs.ServerAddress)
	serverData, ok := g.store.FindServer(addr)
	if !ok {
		// There was no server to be found
		return ErrNoServerFound
	}
	conn.HandshakePacket = pk
	conn.Handshake = hs

	serverData.ConnCh <- conn

	return nil
}

type ServerStore interface {
	FindServer(addr string) (ServerData, bool)
}

type SingleServerStore struct {
	Server ServerData
}

func (store *SingleServerStore) FindServer(addr string) (ServerData, bool) {
	return store.Server, store.Server.ConnCh != nil
}

type DefaultServerStore struct {
	servers map[string]ServerData
}

func (store *DefaultServerStore) FindServer(addr string) (ServerData, bool) {
	server, ok := store.servers[addr]
	if !ok {
		// Client send an invalid address/port; we don't have a v for that address
		return ServerData{}, false
	}
	return server, true
}

func (store *DefaultServerStore) AddServer(addr string, serverData ServerData) {
	store.servers[addr] = serverData
}

func CreateDefaultServerStore() DefaultServerStore {
	store := DefaultServerStore{}
	store.servers = make(map[string]ServerData)
	return store
}
