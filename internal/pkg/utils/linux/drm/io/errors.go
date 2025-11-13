//go:build linux

package io

import "errors"

// ErrNotImplemented is returned by DRM helpers until the actual kernel bindings
// are wired up. The stubs make it possible to sketch higher-level code against
// stable APIs before the syscalls are implemented.
var ErrNotImplemented = errors.New("drm/io: not implemented")
