package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var websocketData = `
<!-- websockets.html -->
<input id="input" type="text" />
<button onclick="send()">Send</button>
<pre id="output"></pre>
<script>
    var input = document.getElementById("input");
    var output = document.getElementById("output");
	var scheme = window.location.protocol == "https:" ? "wss" : "ws";
    var socket = new WebSocket(scheme + "://" + window.location.host + "/echo");

    socket.onopen = function () {
        output.innerHTML += "Status: Connected\n";
    };

    socket.onmessage = function (e) {
        output.innerHTML += "Server: " + e.data + "\n";
    };

    function send() {
        socket.send(input.value);
        input.value = "";
    }
</script>
`

func web() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, websocketData)
	})
}

func ws() {
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		for {
			// Read message from browser
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			// Print the message to the console
			fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			// Write message back to browser
			if err = conn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	})
}

func main() {
	stderr := log.New(os.Stderr, "", 0)
	stdout := log.New(os.Stdout, "", 0)

	args := os.Args[1:]
	processType := "web"
	if len(args) > 0 {
		switch args[0] {
		case "web":
			processType = "web"
		case "ws":
			processType = "ws"
		default:
			stderr.Printf("invalid process type specified: %v\n", args[0])
			os.Exit(1)
		}
	}

	if processType == "web" {
		web()
	} else if processType == "ws" {
		ws()
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	stdout.Printf("listening on :%s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		stderr.Printf("error serving request: %s\n", err.Error())
		os.Exit(1)
	}
}
