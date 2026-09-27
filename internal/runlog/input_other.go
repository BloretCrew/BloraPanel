//go:build !linux && !windows

package runlog

import (
	"context"
	"os"
)

type heldInput struct{}

func newHeldInput() (*heldInput, error)                       { return nil, ErrUnknown }
func (*heldInput) descriptor() (inputDescriptor, error)       { return inputDescriptor{}, ErrUnknown }
func (*heldInput) close()                                     {}
func (*heldInput) write(context.Context, []byte) (int, error) { return 0, ErrUnknown }
func duplicateInput(record) (*os.File, error)                 { return nil, ErrUnknown }
