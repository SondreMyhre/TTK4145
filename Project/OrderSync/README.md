
DETTE ER SKREVET AV AI, KUN BRUK TIL EKSEMPEL TIL FCIS

Functional core example:  
    
    func Step(s State, e Event) (State, []Effect)

Imperative shell example:

    func Run(
    ctx context.Context,
    in <-chan Event,
    out chan<- Msg,
    driver chan<- Cmd,
    // evt. andre avhengigheter
) {
    s := InitState()

    // timers eies her
    var doorTimerC <-chan time.Time

    apply := func(effects []Effect) {
        for _, ef := range effects {
            switch ef.Type {
            case EffectSendMsg:
                out <- ef.Msg
            case EffectDriverCmd:
                driver <- ef.Cmd
            case EffectStartTimer:
                // start/reset timer, sett doorTimerC = timer.C
            }
        }
    }

    for {
        select {
        case <-ctx.Done():
            return

        case ev := <-in:
            s2, effects := Step(s, ev)
            s = s2
            apply(effects)

        case <-doorTimerC:
            s2, effects := Step(s, Event{Type: DoorTimeout})
            s = s2
            apply(effects)
        }
    }
}


DENNE ER GANSKE RYDDIG Å SE PÅ:

https://github.com/angrycompany16/TTK4145-sanntidslab/blob/main/elevalgo/elevator_process.go