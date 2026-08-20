package application

func shouldCollectScanError(err error) bool { return scanErrorGate(err) }
