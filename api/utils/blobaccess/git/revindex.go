package git

import (
	"crypto"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5/plumbing/format/idxfile"
	"github.com/go-git/go-git/v5/plumbing/hash"
	"github.com/mandelsoft/goutils/errors"
)

const (
	revIndexVersion  uint32 = 1
	revIndexHashSHA1 uint32 = 1
)

// revIndexHeader is the magic number of a packfile reverse index.
var revIndexHeader = []byte("RIDX")

// writeRevIndex writes the reverse index of idx: the magic number, the version,
// the hash function, the index position of every object in packfile offset order,
// the packfile checksum, and a checksum over all of the above, in network order
// (see gitformat-pack). go-git v5 can neither write nor name such a file.
//
// OCM v2 stores it next to the packfile, so the archive carries it to stay
// byte-identical across both implementations.
func writeRevIndex(w io.Writer, idx *idxfile.MemoryIndex) error {
	positions, err := revIndexPositions(idx)
	if err != nil {
		return err
	}

	sum := hash.New(crypto.SHA1)
	out := io.MultiWriter(w, sum)

	if _, err := out.Write(revIndexHeader); err != nil {
		return fmt.Errorf("cannot write reverse index header: %w", err)
	}
	if err := binary.Write(out, binary.BigEndian, revIndexVersion); err != nil {
		return fmt.Errorf("cannot write reverse index version: %w", err)
	}
	if err := binary.Write(out, binary.BigEndian, revIndexHashSHA1); err != nil {
		return fmt.Errorf("cannot write reverse index hash function: %w", err)
	}
	if err := binary.Write(out, binary.BigEndian, positions); err != nil {
		return fmt.Errorf("cannot write reverse index entries: %w", err)
	}
	if _, err := out.Write(idx.PackfileChecksum[:]); err != nil {
		return fmt.Errorf("cannot write reverse index packfile checksum: %w", err)
	}
	if _, err := w.Write(sum.Sum(nil)); err != nil {
		return fmt.Errorf("cannot write reverse index checksum: %w", err)
	}
	return nil
}

// revIndexPositions returns the index position of every object, ordered by the
// object's offset in the packfile.
func revIndexPositions(idx *idxfile.MemoryIndex) ([]uint32, error) {
	positions := map[uint64]uint32{}
	if err := eachEntry(idx.Entries, func(position uint32, offset uint64) {
		positions[offset] = position
	}); err != nil {
		return nil, err
	}

	ordered := make([]uint32, 0, len(positions))
	if err := eachEntry(idx.EntriesByOffset, func(_ uint32, offset uint64) {
		ordered = append(ordered, positions[offset])
	}); err != nil {
		return nil, err
	}
	return ordered, nil
}

// eachEntry calls visit for every entry of the iterator with its position in that
// iteration order.
func eachEntry(iter func() (idxfile.EntryIter, error), visit func(position uint32, offset uint64)) error {
	entries, err := iter()
	if err != nil {
		return fmt.Errorf("cannot read packfile index: %w", err)
	}
	defer func() { _ = entries.Close() }()

	for position := uint32(0); ; position++ {
		entry, err := entries.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("cannot read packfile index: %w", err)
		}
		visit(position, entry.Offset)
	}
}
