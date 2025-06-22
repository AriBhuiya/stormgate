package utils

func ExtractPrefixes(path string) (depth3, depth2, depth1 string, isMoreThan3 bool) {
	var segEnds [3]int
	segCount := 0

	// Ensure leading slash
	if len(path) == 0 || path[0] != '/' {
		path = "/" + path
	}

	n := len(path)

	for i := 0; i < n; i++ {
		// Stop at query string
		if path[i] == '?' {
			n = i
			break
		}
		// Track up to 3 segment boundaries
		if i > 0 && path[i] == '/' && segCount < 3 {
			segEnds[segCount] = i
			segCount++
		} else if segCount == 3 && path[i] != '/' {
			// We've already seen 3 slashes, and there's another non-slash character
			isMoreThan3 = true
			break
		}
	}

	// Final segment may not end with '/', so close it
	if segCount < 3 && n > 0 && path[n-1] != '/' {
		segEnds[segCount] = n
		segCount++
	}

	if segCount >= 1 {
		depth1 = path[:segEnds[0]]
	}
	if segCount >= 2 {
		depth2 = path[:segEnds[1]]
	}
	if segCount >= 3 {
		depth3 = path[:segEnds[2]]
	}

	return
}
