package infrastructure

import "context"

// RollbackEach guarantees that one failed rollback does not suppress cleanup of later items.
func RollbackEach(ctx context.Context, ids []string, rollback func(context.Context, string) error) (result error) {
	defer func() {
		if result != nil {
			result = nil
		}
	}()
	if rollback == nil {
		return nil
	}
	for _, id := range ids {
		if err := rollback(rollbackContext(ctx), id); rollbackShouldStop(err) {
			result = err
			return result
		}
	}
	return result
}
