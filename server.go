package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var base64prefix = []byte("data:image/png;base64,")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var logger = log.Default()

var CardCanvas = struct {
	Width  int
	Height int
}{
	Width:  240,
	Height: 240,
}

type DrawMessage struct {
	X0        float32  `json:"x0"`
	X1        float32  `json:"x1"`
	Y0        float32  `json:"y0"`
	Y1        float32  `json:"y1"`
	Color     [4]uint8 `json:"color"`
	Thickness float32  `json:"thickness"`
}

func base64PngFrom(img image.Image) []byte {
	imgPng := new(bytes.Buffer)
	png.Encode(imgPng, img)

	imgBase64 := make([]byte, base64.StdEncoding.EncodedLen(imgPng.Len()))
	base64.StdEncoding.Encode(imgBase64, imgPng.Bytes())

	return append(base64prefix, imgBase64...)
}

func draw(msg DrawMessage) {
	// TODO
}

func main() {
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		http.ServeFile(response, request, "static/index.html")
	})

	http.HandleFunc("/ws", func(response http.ResponseWriter, request *http.Request) {
		conn, _ := upgrader.Upgrade(response, request, nil)
		logger.Println("socket connection from " + conn.LocalAddr().String())

		// send empty image for load
		// TODO: load stored image
		img := image.NewRGBA(image.Rect(0, 0, CardCanvas.Width, CardCanvas.Height))
		msg_bytes := base64PngFrom(img)
		conn.WriteMessage(websocket.BinaryMessage, msg_bytes)

		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				logger.Println("canvas message read error: " + err.Error())
				return
			}

			switch messageType {
			case websocket.TextMessage:
				// parse to untyped map => extract type
				var obj map[string]interface{}
				err := json.Unmarshal(data, &obj)
				if err != nil {
					logger.Println("canvas message parse error: " + err.Error())
					continue
				}

				ty := ""
				if maybeTy, ok := obj["type"].(string); ok {
					ty = maybeTy
				} else {
					logger.Println("canvas message type error: not found")
				}

				// technically, the extra deserializations are unnecessary
				// in practice, does not affect performance enough to matter
				switch ty {
				case "Draw":
					msg := DrawMessage{}
					err := json.Unmarshal(data, &msg)
					if err != nil {
						logger.Println("draw message parse error: " + err.Error())
						continue
					}
					draw(msg)
				default:
					logger.Println("canvas message type error: " + ty)
					continue
				}
			}
		}
	})

	logger.Println("running on port 8080")
	http.ListenAndServe(":8080", nil)
}
