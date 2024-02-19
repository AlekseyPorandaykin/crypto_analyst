package cleaner

import "context"

type Cleaner struct{}

func NewCleaner() *Cleaner {
	return &Cleaner{}
}

func (c *Cleaner) Run(ctx context.Context) error {
	return nil
}
