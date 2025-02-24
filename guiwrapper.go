package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/awesome-gocui/gocui"
)

type guiwrapper struct {
	gui        *gocui.Gui
	messages   []*guimessage
	maxlines   int
	timeformat string
	sync.RWMutex
}

type guimessage struct {
	ts  time.Time
	tag string // TODO assert len=3 ?
	msg string
	// TODO right now only set this when we want to modify tag/highlight/ignore/... later on... not on "system" messages; even though they technically have a sender too...
	nick string
}

func (gw *guiwrapper) formatMessage(gm *guimessage) string {
	formattedDate := gm.ts.Format(gw.timeformat)
	return fmt.Sprintf("[%s]%s%s", formattedDate, gm.tag, gm.msg)
}

func (gw *guiwrapper) redraw() {
	gw.gui.Update(func(g *gocui.Gui) error {
		messageView, err := g.View("messages")
		if err != nil {
			return err
		}

		// If we have reached maxlines and the user is currently reading old lines,
		// i.e. is in an scrolled-up state, do not redraw, because it makes reading
		// messages very hard... Scroll function should make sure to call redraw()
		// when the user is done reading old messages.
		// This introduces some ugly side-effects, e.g. commands typed into chat like
		// /tag will also not be instantly applied when in a scrolled-up state.
		// Or if the user is scrolled up for a very long time, once scrolling down
		// chat will make a big jump redrawing - All of that probably cannot be helped.
		if !messageView.Autoscroll {
			return nil
		}

		gw.Lock()
		defer gw.Unlock()

		// redraw everything
		newbuf := ""
		for _, msg := range gw.messages {
			newbuf += gw.formatMessage(msg) + "\n"
		}

		messageView.Clear()
		fmt.Fprint(messageView, newbuf)

		// When the message view is redrawn, the origin is reset to (0, 0) and an event is added to apply the Autoscroll and set the origin to the bottom of the view.
		// But this event is asynchronous, and not within the guiwrapper.Lock() mutex of this function
		// This can cause race conditions if scrolling fast, i.e. a scroll happens when the origin is still 0, 0 because the view's flush method hasn't been called yet.
		// Manually set the origin here to avoid the race condition on future scrolls
		_, height := messageView.Size()
		messageView.SetOrigin(0, messageView.ViewLinesHeight()-height-1)

		return nil
	})
}

func (gw *guiwrapper) addMessage(m guimessage) {

	gw.Lock()
	// only keep maxlines lines in memory
	delta := len(gw.messages) - gw.maxlines
	if delta > 0 {
		gw.messages = append(gw.messages[delta:], &m)
	} else {
		gw.messages = append(gw.messages, &m)
	}
	gw.Unlock()
	gw.redraw()
}

// currently not in use
func (gw *guiwrapper) applyTag(tag string, nick string) {

	gw.Lock()
	defer gw.Unlock()

	for _, m := range gw.messages {
		if m.nick != "" && strings.EqualFold(m.nick, nick) {
			m.tag = tag
		}
	}
}
