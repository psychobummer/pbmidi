## PBMidi

PsychoBummer's convenience wrapper for [gomidi](gitlab.com/gomidi/midi).

### Why

We have internal usecases where it makes more sense to expose captured MIDI data via a channel reader than through `gomidi`'s hooked callbacks. Maybe you have one, too -- if so, here you go.

### Installation

`go get github.com/psychobummer/pbmidi`

### Example

```golang
func main() {
    // pbmidi.Inputs() will return a []string of all available MIDI inputs
    midiDevice := 0
    stream, err := pbmidi.New(midiDevice)
    if err != nil {
        panic(err)
    }
    defer stream.Stop()

    go func() {
        if err := stream.Start(); err != nil {
            panic(err)
        }
    }

    for midiMsg := range stream.Stream() {
        fmt.Printf("%+v", midiMsg) // { Key: 60, State: 0, Veclocity: 127 }
    }
}
```

