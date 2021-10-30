package jx

import (
	"io"

	"golang.org/x/xerrors"
)

// Elem reads array element and reports whether array has more
// elements to read.
func (d *Decoder) Elem() (ok bool, err error) {
	c, err := d.next()
	if err != nil {
		return false, err
	}
	switch c {
	case '[':
		c, err := d.next()
		if err != nil {
			return false, xerrors.Errorf("next: %w", err)
		}
		if c != ']' {
			d.unread()
			return true, nil
		}
		return false, nil
	case ']':
		return false, nil
	case ',':
		return true, nil
	default:
		return false, xerrors.Errorf(`"[" or "," or "]" expected: %w`, badToken(c))
	}
}

// Arr reads array and calls f on each array element.
func (d *Decoder) Arr(f func(d *Decoder) error) error {
	if err := d.expectNext('['); err != nil {
		return xerrors.Errorf("start: %w", err)
	}
	if err := d.incDepth(); err != nil {
		return xerrors.Errorf("inc: %w", err)
	}
	c, err := d.next()
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	if err != nil {
		return err
	}
	if c == ']' {
		return d.decDepth()
	}
	d.unread()
	if err := f(d); err != nil {
		return xerrors.Errorf("callback: %w", err)
	}

	c, err = d.next()
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	if err != nil {
		return xerrors.Errorf("next: %w", err)
	}
	for c == ',' {
		if err := f(d); err != nil {
			return xerrors.Errorf("callback: %w", err)
		}
		if c, err = d.next(); err != nil {
			return xerrors.Errorf("next: %w", err)
		}
	}
	if c != ']' {
		return xerrors.Errorf("end: %w", badToken(c))
	}
	return d.decDepth()
}
