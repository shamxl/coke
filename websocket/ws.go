package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type Args struct {
  Ws *Ws
  Message []byte
  Err error
  Reason string
}
type onEventFunc func (*Args)
type Ws struct {
  Interrupt chan any
  Closed chan any
  Response *http.Response

  wsUrl string
  dailer *websocket.Conn
  onOpen onEventFunc
  onClose onEventFunc
  onMessage onEventFunc
  onError onEventFunc
}

func NewWebsocket (url string) *Ws {
  var ws *Ws = &Ws{
    wsUrl: url,
  }
  var response *http.Response
  var err error 
  var dailer *websocket.Conn
  dailer, response, err = websocket.DefaultDialer.Dial(url, nil)
  if err != nil {
    ws.onError(&Args{
      Err: err,
    })
  }

  ws.Response = response
  ws.dailer = dailer
  return ws
}

func (ws *Ws) SetOpen (fn onEventFunc) {
  ws.onOpen = fn
}

func (ws *Ws) SetClose (fn onEventFunc) {
  ws.onClose = fn
}

func (ws *Ws) SetMessage (fn onEventFunc) {
  ws.onMessage = fn
}

func (ws *Ws) SetError (fn onEventFunc) {
  ws.onError = fn
}

func (ws *Ws) Send (data []byte) {
  var err error = ws.dailer.WriteMessage(websocket.BinaryMessage, data)
  if err != nil {
    ws.onError (&Args{
      Err: err,
    })
  }
}

func (ws *Ws) Close (reason string) {
  close(ws.Closed)
  ws.onClose (&Args{
    Reason: reason,
  })
}

