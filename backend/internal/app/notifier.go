package app

import "github.com/xcreativs/gigmann/internal/ports"

// fanoutNotifier forwards each event to several notifiers (e.g. the realtime hub
// and the critical-push sweep), so one broadcast signal drives both channels.
type fanoutNotifier struct{ targets []ports.Notifier }

// FanoutNotifier returns a ports.Notifier that forwards every event to all targets.
func FanoutNotifier(targets ...ports.Notifier) ports.Notifier {
	return fanoutNotifier{targets: targets}
}

func (f fanoutNotifier) Notify(event string) {
	for _, t := range f.targets {
		t.Notify(event)
	}
}
