package infrastructure

import "context"

// RollbackEach guarantees that one failed rollback does not suppress cleanup of later items.
func RollbackEach(ctx context.Context, ids []string, rollback func(context.Context, string) error) error {
	for _, id := range ids {
		if err := rollback(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
