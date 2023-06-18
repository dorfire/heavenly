package pkga

import (
	"github.com/samber/lo"
	"log"
)

func DoSomething() error {
	p := lo.ToPtr(1)
	log.Printf("Hello world %d", *p)
	return nil
}
