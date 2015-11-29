package main

import (
	"./lib/"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	NewLightifyServer("192.168.42.147:4000")
	//NewLightifyServer(":4000")
}

func lightsHandler(w http.ResponseWriter, r *http.Request) {

	message := new(golightify.LightifyRequest_ListAllLightsReq)
	message.AllDetails = 1 // 0 = returns only names but currently response not handled

	response := golightify.SendLightifyRequest(message)

	responseJSON, _ := json.Marshal(response)

	fmt.Fprintf(w, string(responseJSON))
}

func groupsHandler(w http.ResponseWriter, r *http.Request) {

	message := new(golightify.LightifyRequest_ListAllGroupsReq)
	message.U1 = 1
	message.U2 = 0

	response := golightify.SendLightifyRequest(message)

	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprintf(w, string(responseJSON))
}

type lightHandler struct{}

func (h lightHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	fmt.Println(urlParts)
	if len(urlParts) == 0 {
		http.NotFound(w, r)
		return
	}

	id, err := hex.DecodeString(urlParts[0])
	if err != nil || len(id) != 8 {
		http.NotFound(w, r)
		return
	}
	if len(urlParts) == 1 {
		message := new(golightify.LightifyRequest_LightDetailsReq)
		copy(message.Id[:], id[:])
		response := golightify.SendLightifyRequest(message)
		responseJSON, err := json.Marshal(response)
		if err != nil {
			log.Println(err)
		}
		fmt.Fprintf(w, string(responseJSON))
	} else {
		r.ParseForm() // Parses the request body

		var message golightify.LightifyRequest
		switch strings.ToLower(urlParts[1]) {
		case "onoff":

			req := new(golightify.LightifyRequest_LightOnOffReq)
			switch strings.ToLower(r.Form.Get("state")) {
			case "on":
				req.On = 1
			case "off":
				req.On = 0
			default:
				http.NotFound(w, r)
				return
			}
			copy(req.Id[:], id[:])
			message = req
		case "bri":
			req := new(golightify.LightifyRequest_LightBrightnessReq)

			if i, err := strconv.Atoi(r.Form.Get("bri")); err == nil {
				req.Bri = byte(i)
			} else {
				http.NotFound(w, r)
				return
			}

			if i, err := strconv.Atoi(r.Form.Get("time")); err == nil {
				req.Time = uint16(i)
			} else {
				req.Time = 0
			}

			copy(req.Id[:], id[:])
			message = req

		case "temp":
			req := new(golightify.LightifyRequest_LightTemperatureReq)

			if i, err := strconv.Atoi(r.Form.Get("temp")); err == nil {
				req.Temp = uint16(i)
			} else {
				req.Temp = 0
			}

			if i, err := strconv.Atoi(r.Form.Get("time")); err == nil {
				req.Time = uint16(i)
			} else {
				req.Time = 0
			}

			copy(req.Id[:], id[:])
			message = req

		case "color":
			req := new(golightify.LightifyRequest_LightColorReq)

			if i, err := strconv.Atoi(r.Form.Get("r")); err == nil {
				req.Color.Red = byte(i)
			} else {
				req.Color.Red = 0
			}

			if i, err := strconv.Atoi(r.Form.Get("g")); err == nil {
				req.Color.Green = byte(i)
			} else {
				req.Color.Green = 0
			}

			if i, err := strconv.Atoi(r.Form.Get("b")); err == nil {
				req.Color.Blue = byte(i)
			} else {
				req.Color.Blue = 0
			}

			if i, err := strconv.Atoi(r.Form.Get("x")); err == nil {
				req.U1 = byte(i)
			} else {
				req.U1 = 0xff
			}

			if i, err := strconv.Atoi(r.Form.Get("time")); err == nil {
				req.Time = uint16(i)
			} else {
				req.Time = 0
			}

			copy(req.Id[:], id[:])
			message = req

		default:
			http.NotFound(w, r)
			return
		}
		if message != nil {
			log.Printf("request: %+v", message)
			response := golightify.SendLightifyRequest(message)
			responseJSON, err := json.Marshal(response)
			if err != nil {
				log.Println(err)
			}
			fmt.Fprintf(w, string(responseJSON))
		}
	}
}

type groupHandler struct{}

func (h groupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	fmt.Println(urlParts)
	if len(urlParts) == 0 {
		http.NotFound(w, r)
		return
	}

	if id, err := strconv.Atoi(urlParts[0]); err == nil && id != 0 {
		message := new(golightify.LightifyRequest_GroupDetailsReq)
		message.Id = golightify.LightifyGroupId(id)
		response := golightify.SendLightifyRequest(message)
		responseJSON, err := json.Marshal(response)
		if err != nil {
			log.Println(err)
		}
		fmt.Fprintf(w, string(responseJSON))
	}
}

func NewLightifyServer(lightifyAddr string) error {

	err := golightify.NewLightifyBridge(lightifyAddr)
	if err != nil {
		log.Printf("Connection to lightify bridge failed: ", err.Error())
		return err
	}

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	mux.HandleFunc("/api/lights", lightsHandler)
	mux.HandleFunc("/api/groups", groupsHandler)
	mux.Handle("/api/lights/", http.StripPrefix("/api/lights/", lightHandler{}))
	mux.Handle("/api/groups/", http.StripPrefix("/api/groups/", groupHandler{}))

	log.Println("Listening... ", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Printf("Server failed: ", err.Error())
		return err
	}

	return nil
}
