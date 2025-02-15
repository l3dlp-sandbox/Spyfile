package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	_ "github.com/qodrorid/godaemon"
)

type Events struct {
	Event   chan Event
	NoEvent chan NoEvent
}

type Event struct {
	Name string
}

type NoEvent struct {
}

type Error struct {
	Error error
}

func atime(fi os.FileInfo) time.Time {
	return time.Unix(0, fi.Sys().(*syscall.Stat_t).Atim.Nsec)
}

func main() {
	args := os.Args
	var file string
	if len(args) < 2 {
		file = "/datamix/config"
	} else {
		file = args[1]
	}

	ch := make(chan any)
	done := make(chan bool)
	go func() {
		var latest_updated time.Time
		var latest_size int64
		f, err := os.Stat(file)
		if err != nil {
			close(ch)
			log.Println(err)
			return

		}

		atime1 := atime(f)

		latest_updated = atime1
		latest_size = f.Size()
		for {

			f_n, err := os.Stat(file)
			if err != nil {
				ch <- Error{
					Error: err,
				}
				continue
			}

			atime := atime(f_n)
			if atime.Format("2006-01-02 15:04:05") != latest_updated.Format("2006-01-02 15:04:05") {
				latest_updated = atime
				if latest_size == f_n.Size() {
					ch <- Event{
						Name: f.Name() + ": OPENED",
					}
					continue
				} else if latest_size != f_n.Size() {
					latest_size = f_n.Size()
					ch <- Event{
						Name: f.Name() + ": WRITE",
					}
					continue
				} else {
					ch <- NoEvent{}
					continue
				}
			} else {
				ch <- NoEvent{}
				continue
			}
		}
	}()

	go func() {
		for {
			select {
			case evt, ok := <-ch:
				_ = ok
				switch evt.(type) {
				case Event:
					fmt.Println("EVENT:", evt.(Event).Name)
					continue
				case Error:
					/*
						err := beeep.Alert("Spyfile", evt.(Error).Error.Error(), "logo.jpg")
						if err != nil {
							log.Println(err)
						}
					*/
					fmt.Println("ERROR:", evt.(Error).Error.Error())
					continue
				}
			}
		}
	}()

	<-done
}
