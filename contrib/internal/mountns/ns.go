package mountns

import (
	"fmt"
	"runtime"
	"sync"

	"golang.org/x/sys/unix"
)

// Executes runs the closure in a new mount namespace.
//
// NOTE: The caller should not call runtime.UnlockOSThread or fork any new
// goroutines, because it's risk. The thread in the new mount namespace should
// be cleanup by Go runtime when it exits without unlock OS thread.
func Executes(run func() error) error {
	var wg sync.WaitGroup
	wg.Add(1)

	var innerErr error
	go func() (retErr error) {
		defer wg.Done()

		defer func() {
			innerErr = retErr
		}()

		runtime.LockOSThread()

		err := unix.Unshare(unix.CLONE_FS | unix.CLONE_NEWNS)
		if err != nil {
			return fmt.Errorf("failed to create a new mount namespace: %w", err)
		}
		return run()
	}()
	wg.Wait()

	return innerErr
}
