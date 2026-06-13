package capture

// ActiveSweepCount returns the number of device entries in the sweeps map.
// It exists for tests that verify completed sweeps are removed.
func (c *Controller) ActiveSweepCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.sweeps)
}
