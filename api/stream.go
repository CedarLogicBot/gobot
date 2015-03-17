package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hybridgroup/gobot"
)

// Used to implement Server-Sent Events
type Stream struct {
	api      *API
	dataChan chan string
}

func NewStream(api *API) *Stream {
	s := &Stream{api, make(chan string)}
	return s
}

func (s *Stream) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	a := s.api
	f, _ := res.(http.Flusher)
	c, _ := res.(http.CloseNotifier)

	closer := c.CloseNotify()

	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")

	if event := a.gobot.Robot(req.URL.Query().Get(":robot")).
		Device(req.URL.Query().Get(":device")).(gobot.Eventer).
		Event(req.URL.Query().Get(":event")); event != nil {
		gobot.On(event, func(data interface{}) {
			d, _ := json.Marshal(data)
			s.dataChan <- string(d)
		})

		for {
			select {
			case data := <-s.dataChan:
				fmt.Fprintf(res, "data: %v\n\n", data)
				f.Flush()
			case <-closer:
				log.Println("Closing connection")
				return
			}
		}
	} else {
		a.writeJSON(map[string]interface{}{
			"error": "No Event found with the name " + req.URL.Query().Get(":event"),
		}, res)
	}
}
