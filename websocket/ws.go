package websocket

import (
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type Args struct {
	Ws      *Ws
	Message []byte
	Err     error
	Reason  string
}
type onEventFunc func(*Args)
type Ws struct {
	Interrupt chan os.Signal
	Closed    chan any
	Response  *http.Response
	Wg        *sync.WaitGroup

	wsUrl     string
	dailer    *websocket.Conn
	onOpen    onEventFunc
	onClose   onEventFunc
	onMessage onEventFunc
	onError   onEventFunc
}

func NewWebsocket(url string) *Ws {
	var ws *Ws = &Ws{
		wsUrl:     url,
		Wg:        &sync.WaitGroup{},
		Interrupt: make(chan os.Signal),
		Closed:    make(chan any),
	}
	var response *http.Response
	var err error
	var dailer *websocket.Conn
	dailer, response, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		ws.onError(&Args{
			Ws:  ws,
			Err: err,
		})
	}

	ws.Response = response
	ws.dailer = dailer
	return ws
}

func (ws *Ws) SetOpen(fn onEventFunc) {
	ws.onOpen = fn
}

func (ws *Ws) SetClose(fn onEventFunc) {
	ws.onClose = fn
}

func (ws *Ws) SetMessage(fn onEventFunc) {
	ws.onMessage = fn
}

func (ws *Ws) SetError(fn onEventFunc) {
	ws.onError = fn
}

func (ws *Ws) openCb(args *Args) {
	if ws.onOpen != nil {
		ws.onOpen(args)
	}
}

func (ws *Ws) closeCb(args *Args) {
	if ws.onClose != nil {
		ws.onClose(args)
	}
}

func (ws *Ws) messageCb(args *Args) {
	if ws.onMessage != nil {
		ws.onMessage(args)
	}
}

func (ws *Ws) errorCb(args *Args) {
	if ws.onError != nil {
		ws.onError(args)
	}
}

func (ws *Ws) Send(data []byte) {
	var err error = ws.dailer.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		ws.errorCb(&Args{
			Ws:  ws,
			Err: err,
		})
	}
}

func (ws *Ws) Close(reason string) {
	close(ws.Closed)
	ws.closeCb(&Args{
		Ws:     ws,
		Reason: reason,
	})
	ws.Wg.Wait()
}

func (ws *Ws) Connect() {
	ws.openCb(&Args{
		Ws: ws,
	})

	ws.Wg.Add(1)
	go func() {
		for {
			select {
			case <-ws.Closed:
				ws.Wg.Done()
				return

			case <-ws.Interrupt:
				ws.Wg.Done()
				return

			default:
				_, msg, err := ws.dailer.ReadMessage()
				if err != nil {
					ws.errorCb(&Args{
						Ws:  ws,
						Err: err,
					})
					ws.Wg.Done()
					return
				}

				if msg != nil {
					ws.messageCb(&Args{
						Ws:      ws,
						Message: msg,
					})
				}
			}
		}
	}()
}
