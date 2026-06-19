package url

import "context"

func (s *service) DeleteByExpireAt(ctx context.Context) error {
	return s.urlRepo.DeleteByExpireAt(ctx)
}
