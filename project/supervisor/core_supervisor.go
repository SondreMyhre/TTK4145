package supervisor


import (
	
)

func ShouldRestart(policy RestartPolicy, err error) bool {
	switch policy {
	case Permanent:
		return true
	case Transient:
		return err != nil
	case Temporary:
		return false
	default:
		return err != nil
	}
}