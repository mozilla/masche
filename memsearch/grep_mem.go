package memsearch

//TODO(aemartinez): Add documentation.

type MemoryRegion struct {
	address uintptr
	size    uint
}

type Process interface {
	Close() error
	NextReadableMemoryRegion(address uintptr) (MemoryRegion, error)
	ReadMemory(address uintptr, size uint) ([]byte, error)
	CopyMemory(address uintptr, buffer []byte) error
}

// Find looks for needle in Process p's memory.
// It works like FindNext but it doesn't search the memory in a linear way.
// The address returned is not guaranteed to be the lowest address that contains the needle.
func Find(p Process, needle []byte) (addr uintptr, found bool, errs []error) {
	type result struct {
		address uintptr
		found   bool
		err     error
	}

	workers := 50

	results := make(chan result, workers)

	// regions chan is used for sending jobs to the workers.
	// Each job consists in one memory region to search in.
	regions := make(chan MemoryRegion, workers)

	end := make(chan bool) // This chan will be closed to stop all workers after we find the first result.

	// spawn workers
	for i := 0; i < workers; i++ {
		go func() {
			for r := range regions {
				addr, found, err := findInRegion(p, r, needle)
				select {
				case results <- result{addr, found, err}:
				case <-end:
					return
				}
			}
		}()
	}

	// iterate over all regions
	count := 0
	go func() {
		defer close(regions)

		address := uintptr(0)
		for {
			region, err := p.NextReadableMemoryRegion(address)
			if err != nil {
				//TODO(mvanotti): return error.
				return
			}
			if region.size == 0 {
				return
			}
			select {
			case regions <- region:
				count += 1
			case <-end:
				return
			}
			address = region.address + uintptr(region.size)
		}
	}()

	// check for results.
	found = false
	addr = 0
	for done := 0; done < count; done++ {
		r := <-results
		if r.err != nil {
			errs = append(errs, r.err)
		}
		if r.found {
			found = r.found
			addr = r.address
			break
		}
	}

	// we don't want to keep the workers waiting so we let them know that we are done by closing this channel.
	close(end)

	return addr, found, errs
}

// FindNext finds for the first occurrence of needle in the memory of Process ph after the given address.
func FindNext(ph Process, address uintptr, needle []byte) (uintptr, bool, error) {
	region, err := ph.NextReadableMemoryRegion(address)
	if err != nil {
		return 0, false, err
	}
	for region.address != 0 {
		res, found, err := findInRegion(ph, region, needle)
		if err != nil {
			return 0, false, err
		} else if found {
			return res, found, nil
		}
		region, err = ph.NextReadableMemoryRegion(region.address + uintptr(region.size))
		if err != nil {
			return 0, false, err
		}
	}
	return 0, false, nil
}

// findInRegion looks for the needle inside a given memory region.
func findInRegion(p Process, region MemoryRegion, needle []byte) (uintptr, bool, error) {
	//TODO: We should change this for a more efficient algorithm.

	buf := make([]byte, len(needle)) //TODO: Use a bigger buffer.
	for i := uint(0); i < region.size-uint(len(buf)); i++ {
		err := p.CopyMemory(region.address+uintptr(i), buf)
		if err != nil {
			return 0, false, err
		}
		if areEqual(buf, needle) {
			return region.address + uintptr(i), true, err
		}
	}
	return 0, false, nil
}

// areEqual returns true if and only if the two slices contains te same elements.
func areEqual(s1 []byte, s2 []byte) bool {
	for index, _ := range s1 {
		if s1[index] != s2[index] {
			return false
		}
	}
	return true
}
