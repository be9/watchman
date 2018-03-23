package watchman

import (
	"context"
	"encoding/json"
)

type object map[string]interface{}

type eventloop struct {
	commands    chan<- string
	results     <-chan object
	unilaterals <-chan object
}

func loop(ctx context.Context, s *server) (l *eventloop) {
	commands := make(chan string)
	results := make(chan object)
	unilaterals := make(chan object)
	l = &eventloop{
		commands:    commands,
		results:     results,
		unilaterals: unilaterals,
	}

	expectCommand := func() (ok bool) {
		for {
			select {
			case command, ok := <-commands:
				if ok {
					s.commands <- command
				}
				return ok
			case pdu, ok := <-s.events:
				if ok {
					var event object
					if err := json.Unmarshal(pdu, &event); err != nil {
						ok = false
						event = object{"error": err.Error()}
					}
					unilaterals <- event
				}
				return ok
			case <-ctx.Done():
				return false
			}
		}
	}

	expectResult := func() (ok bool) {
		for {
			select {
			case pdu, ok := <-s.events:
				if ok {
					var event object
					if err := json.Unmarshal(pdu, &event); err != nil {
						ok = false
						event = object{"error": err.Error()}
					}
					if _, u8l := event["log"]; u8l {
						unilaterals <- event
					} else if _, u8l := event["subscription"]; u8l {
						unilaterals <- event
					} else {
						results <- event
					}
				}
				return ok
			case <-ctx.Done():
				return false
			}
		}
	}

	go func() {
		defer close(commands)
		defer close(results)
		defer close(unilaterals)
		for {
			if ok := expectCommand(); !ok {
				return
			}
			if ok := expectResult(); !ok {
				return
			}
		}
	}()

	return
}
