package pbmidi

import (
	"context"
	"fmt"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
)

// NoteState an iota for mapping key up/key down
type NoteState uint8

// TODO not sure if this is the most friendly?
const (
	NoteOn = NoteState(iota)
	NoteOff
)

// Message is a simple go struct representation of a midi event. Basically a wrapper around the upstream gomidi/midi data.
type Message struct {
	Key       uint8
	State     NoteState
	Veclocity uint8
}

// MidiStream defines a simple interface for receiving MIDI messages as a stream of data.
type MidiStream interface {
	Start() error
	Stop()
	Stream() <-chan Message
}

// PBMidi provides a means of receiving midi messages from a machine-local source,
// ie: a keyboard or DAW connected to the same MIDI bus.
type PBMidi struct {
	ctx    context.Context
	cancel context.CancelFunc
	driver *rtmididrv.Driver
	input  midi.In
	stream chan Message
}

// New returns a new *PBMidi, with the underlying midi driver started
// and inputs initialized.
func New(deviceNum int) (MidiStream, error) {
	driver, err := rtmididrv.New()
	if err != nil {
		return nil, err
	}

	inputs, err := driver.Ins()
	if err != nil {
		return nil, err
	}

	if deviceNum > len(inputs) {
		return nil, fmt.Errorf("requested device %d, but only have %d devices", deviceNum, len(inputs))
	}

	input := inputs[0] // this is probably sufficient for now.
	input.Open()

	ctx, cancel := context.WithCancel(context.Background())
	pbMidi := PBMidi{
		ctx:    ctx,
		cancel: cancel,
		driver: driver,
		input:  input,
		stream: make(chan Message),
	}
	return &pbMidi, nil
}

// Inputs will return a []string{} with all the available midi inputs
func Inputs() ([]string, error) {
	driver, err := rtmididrv.New()
	if err != nil {
		return nil, err
	}

	inputs, err := driver.Ins()
	if err != nil {
		return nil, err
	}

	available := make([]string, len(inputs))
	for i, name := range inputs {
		available[i] = fmt.Sprintf("%s", name)
	}

	return nil, err
}

// Stop shuts down the underlying midi driver and input
// TOOD: do we care about errors here? Maybe just log?
func (p *PBMidi) Stop() {
	p.input.StopListening()
	p.driver.Close()
	p.cancel()
}

// Start starts the process which will convert received midi messages into more friendly Message structs.
// It's a blocking process so you'll want to start this in a go func. Calling `PBMidi.Stop()`` will
// cause this process to shutdown and close the `PBMidi.Stream` channel.
// Returns error on error.
func (p *PBMidi) Start() error {
	r := reader.New(
		reader.NoLogger(),
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			message := makeMessage(msg)
			select {
			case <-p.ctx.Done():
				close(p.stream)
			case p.stream <- message:
			}
		}),
	)
	return r.ListenTo(p.input)
}

// Stream returns a channel that you can read messages off of
func (p *PBMidi) Stream() <-chan Message {
	return p.stream
}

func makeMessage(msg midi.Message) Message {
	var message Message
	switch v := msg.(type) {
	case channel.NoteOn:
		message = Message{
			State:     NoteOn,
			Key:       v.Key(),
			Veclocity: v.Velocity(),
		}
	case channel.NoteOff:
		message = Message{
			State: NoteOff,
			Key:   v.Key(),
		}
	}
	return message
}
