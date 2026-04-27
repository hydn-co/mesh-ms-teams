package helpers

import "fmt"

// CheckInitialized returns an error if the feature has not been initialized.
// This should be called in Start and Stop methods to ensure Init was called first.
func CheckInitialized(initialized bool) error {
	if !initialized {
		return fmt.Errorf("feature not initialized; call Init first")
	}
	return nil
}
