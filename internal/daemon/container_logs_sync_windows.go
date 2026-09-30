package daemon

// Windows cannot FlushFileBuffers on a directory handle opened by os.Open.
// The archive file is synced before rename; directory power-loss durability
// is not claimed, matching the other Windows archive adapters.
func syncContainerLogDirectory(string) error { return nil }
