package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ImageData struct {
	Height int32  `json:"height"`
	Width  int32  `json:"width"`
	Mime   string `json:"mime"`
	Name   string `json:"name"`
	Url    string `json:"url"`
}
type Emote struct {
	Prefix string      `json:"prefix"`
	Twitch bool        `json:"twitch"`
	Theme  string      `json:"theme"`
	Image  []ImageData `json:"image"`
}

type sggEndpointResponse struct {
	Default []string `json:"default"`
}

func getEmoteNames(e []Emote) []string {
	s := make([]string, len(e))
	for i, em := range e {
		s[i] = em.Prefix
	}
	return s
}

func getEmotes() ([]string, error) {
	emotes := make([]string, 0)
	var emoteEndpoint string

	switch chatServer {
	case "dgg":
		emoteEndpoint = "https://cdn.destiny.gg/emotes/emotes.json"
	case "ogg":
		emoteEndpoint = "https://cdn.omniliberal.dev/emotes/emotes.json"
	case "sgg":
		emoteEndpoint = "https://raw.githubusercontent.com/MemeLabs/chat-gui/master/assets/emotes.json"
	}
	resp, err := http.Get(emoteEndpoint)
	if err != nil {
		return emotes, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return emotes, fmt.Errorf("emote endpoint status code %d", resp.StatusCode)
	}

	switch chatServer {
	case "sgg":
		var em sggEndpointResponse
		if err := json.NewDecoder(resp.Body).Decode(&em); err != nil {
			return emotes, err
		}
		emotes = append(emotes, em.Default...)
	default:
		var em []Emote
		if err := json.NewDecoder(resp.Body).Decode(&em); err != nil {
			return emotes, err
		}
		emotes = append(emotes, getEmoteNames(em)...)
	}

	return emotes, nil
}
