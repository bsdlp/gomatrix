package main

import (
	"github.com/nsf/termbox-go"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Stream updates a StreamDisplay with new data updates
type Stream struct {
	display  *StreamDisplay
	speed    int
	length   int
	headPos  int
	tailPos  int
	stopCh   chan bool
	headDone bool
}

func (s *Stream) run() {
	var lastRune rune
	for {
		select {
		case <-s.stopCh:
			log.Printf("Stream on SD %d was stopped.\n", s.display.column)
			goto done
		case <-time.After(time.Duration(s.speed) * time.Millisecond):
			// add a new rune if there is space in the stream
			if !s.headDone && s.headPos <= curSizes.height {
				newRune := characters[rand.Intn(len(characters))]
				termbox.SetCell(s.display.column, s.headPos-1, lastRune, termbox.ColorGreen, termbox.ColorBlack)
				termbox.SetCell(s.display.column, s.headPos, newRune, termbox.ColorWhite, termbox.ColorBlack)
				lastRune = newRune
				s.headPos++
			} else {
				s.headDone = true
			}

			// clear rune at the tail of the stream
			if s.tailPos > 0 || s.headPos >= s.length {
				if s.tailPos == 0 {
					// tail is being incremented for the first time. there is space for a new stream
					s.display.newStream <- true
				}
				if s.tailPos < curSizes.height {
					termbox.SetCell(s.display.column, s.tailPos, ' ', termbox.ColorBlack, termbox.ColorBlack) //'\uFF60'
					s.tailPos++
				} else {
					goto done
				}
			}
		}
	}
done:
	delete(s.display.streams, s)
}

// StreamDisplay represents a horizontal line in the terminal on which `Stream`s are displayed.
// StreamDisplay also creates the Streams themselves
type StreamDisplay struct {
	column      int
	stopCh      chan bool
	streams     map[*Stream]bool
	streamsLock sync.Mutex
	newStream   chan bool
}

func (sd *StreamDisplay) run() {
	for {
		select {
		case <-sd.stopCh:
			// lock this SD forever
			sd.streamsLock.Lock()

			// stop streams for this SD
			for s, _ := range sd.streams {
				s.stopCh <- true
			}

			// log that SD has closed
			log.Printf("StreamDisplay on column %d stopped.\n", sd.column)

			// close this goroutine
			return

		case <-sd.newStream:
			// have some wait before the first stream starts..
			// <-time.After(time.Duration(rand.Intn(9000)) * time.Millisecond) //++ TODO: .After or .Sleep??
			time.Sleep(time.Duration(rand.Intn(9000)) * time.Millisecond)

			// lock map
			sd.streamsLock.Lock()

			// crekate new stream instance
			s := &Stream{
				display: sd,
				stopCh:  make(chan bool),
				speed:   30 + rand.Intn(110),
				length:  6 + rand.Intn(6), // length of a stream is between 6 and 12 runes
			}

			// store in streams map
			sd.streams[s] = true

			// run the stream in a goroutine
			go s.run()

			// unlock map
			sd.streamsLock.Unlock()
		}
	}
}
