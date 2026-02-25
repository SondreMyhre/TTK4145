package supervisor

import (
	"context"
	"time"
)

func (supervisor Supervisor) Run(ctx context.Context) error {
	for {
		err := supervisor.runWorker(ctx)

		// graceful shutdown
		if ctx.Err() != nil {
			return nil
		}
		// decide if we want to STOP
		switch supervisor.Child.Restart {
			case Permanent:
			// always restart, dont do anything

			case Transient:
				if err == nil {
					return nil //clean exit
				}
			case Temporary:
				//never restart
				return err

			default:
				//treat unknown as transient
				if err == nil {
					return nil
				}
		}

		// restart delay
		if supervisor.RestartDelay > 0 {
			select {
			case <-time.After(supervisor.RestartDelay):
			case <-ctx.Done():
				return nil
			}
		}
	}
}
