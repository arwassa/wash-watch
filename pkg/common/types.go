package utils

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

type Machine struct {
	Name       string
	StateField string
}

func (machine *Machine) State() string {
	//"Idle until"
	//"Busy"
	//"Time left"
	//"Closing"
	closing := regexp.MustCompile("Closing (.+)")
	timeLeft := regexp.MustCompile("Time left: (.+)")
	idleUntil := regexp.MustCompile("Idle until (.+)")
	closedUntil := regexp.MustCompile("Closed until (.+)")
	// washer[2] reserved from 15:30
	if timeLeft.MatchString(machine.StateField) {
		return "busy"
	}
	if idleUntil.MatchString(machine.StateField) {
		return "idle"
	}
	if closing.MatchString(machine.StateField) {
		return "closing"
	}
	if closedUntil.MatchString(machine.StateField) {
		return "closed"
	}
	return strings.ToLower(machine.StateField)
}

func (machine *Machine) TimeLeft() time.Duration {
	if machine.State() != "busy" {
		return time.Duration(0)
	}
	timeLeft := regexp.MustCompile("Time left: (.+)")
	matches := timeLeft.FindStringSubmatch(machine.StateField)
	if len(matches) > 0 {
		d, e := time.ParseDuration(strings.ReplaceAll(matches[1], " min.", "m"))
		if e != nil {
			log.Fatal(e)
		}
		return d
	}
	return time.Minute
}

func (machine *Machine) IdleUntil() time.Time {
	if machine.State() != "idle" {
		return time.Now()
	}
	idleUntil := regexp.MustCompile("Idle until (.+)")
	matches := idleUntil.FindStringSubmatch(machine.StateField)
	if len(matches) > 0 {
		d, e := time.ParseInLocation("15:04", matches[1], time.Local)
		if e != nil {
			log.Fatal(e)
		}
		return d
	}
	return time.Now()
}

func (machine *Machine) ClosedUntil() time.Time {
	if machine.State() != "closed" {
		return time.Now()
	}
	idleUntil := regexp.MustCompile("Closed until (.+)")
	matches := idleUntil.FindStringSubmatch(machine.StateField)
	if len(matches) > 0 {
		d, e := time.ParseInLocation("15:04", matches[1], time.Local)
		if e != nil {
			log.Fatal(e)
		}
		return d
	}
	return time.Now()
}

func (machine *Machine) Type() string {
	washMatch := regexp.MustCompile("Vask (\\d+)")
	if washMatch.Match([]byte(machine.Name)) {
		return "washer"
	}
	return "dryer"
}

func (machine *Machine) Id() string {
	idMatch := regexp.MustCompile("\\d+")
	return idMatch.FindString(machine.Name)
}

func (machine Machine) String() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("%s[%s]", machine.Type(), machine.Id()))
	state := machine.State()
	if state == "busy" {
		d := machine.TimeLeft()
		if d > 0 {
			s.WriteString(fmt.Sprintf(" busy for %s", d.String()))
		} else {
			s.WriteString(fmt.Sprintf(" busy"))
		}
	} else if state == "idle" {
		t := machine.IdleUntil()
		s.WriteString(fmt.Sprintf(" idle until %s", t.Format("15:04")))
	} else if state == "closed" {
		t := machine.ClosedUntil()
		s.WriteString(fmt.Sprintf(" closed until %s", t.Format("15:04")))
	} else {
		s.WriteString(fmt.Sprintf(" %s", state))
	}
	return s.String()
}
