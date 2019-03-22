package dictionary

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sync"
)

// FileVer is the current compatibility level of saved Files.
const FileVer = "DICT2\x00"

// File implements an efficient Store which is faster to initialize
// and uses a lot less memory (~15 MB total) than WordMap.
type File struct {
	mu  sync.Mutex
	idx map[string]int64
	df  interface {
		io.Reader
		io.Seeker
		io.Closer
	}
}

// CreateFile exports a WordMap to a file. The files specified will
// be overwritten if they exist.
func CreateFile(wm WordMap, idxf, dbf string) error {
	f, err := os.Create(dbf)
	if err != nil {
		return fmt.Errorf("could not create db: %v", err)
	}
	defer f.Sync()
	defer f.Close()

	_, err = f.WriteString(FileVer)
	if err != nil {
		return fmt.Errorf("could not write version string: %v", err)
	}

	idx := map[string]int64{}
	revidx := map[*Word]int64{}
	buf := new(bytes.Buffer)
	for k, v := range wm {
		if cur, ok := revidx[v]; ok {
			idx[k] = cur
			continue
		}

		buf.Reset()

		cur, err := f.Seek(0, 1)
		if err != nil {
			return fmt.Errorf("could not get current file offset: %v", err)
		}

		zw, err := zlib.NewWriterLevel(buf, zlib.BestCompression)
		if err != nil {
			return fmt.Errorf("could not compress word: %v", err)
		}

		err = gob.NewEncoder(zw).Encode(v)
		if err != nil {
			return fmt.Errorf("could not encode word: %v", err)
		}

		zw.Close()

		err = binary.Write(f, binary.LittleEndian, int64(buf.Len()))
		if err != nil {
			return fmt.Errorf("could not write word size: %v", err)
		}

		_, err = f.Write(buf.Bytes())
		if err != nil {
			return fmt.Errorf("could not write word: %v", err)
		}

		idx[k] = cur
		revidx[v] = cur
	}

	df, err := os.Create(idxf)
	if err != nil {
		return fmt.Errorf("could not create idx: %v", err)
	}
	defer df.Sync()
	defer df.Close()

	_, err = df.WriteString(FileVer)
	if err != nil {
		return fmt.Errorf("could not write version string: %v", err)
	}

	zw, err := zlib.NewWriterLevel(df, zlib.BestCompression)
	if err != nil {
		return fmt.Errorf("could not compress idx: %v", err)
	}
	defer zw.Close()

	err = gob.NewEncoder(zw).Encode(idx)
	if err != nil {
		return fmt.Errorf("could not encode idx: %v", err)
	}

	return nil
}

// OpenFile opens a dictionary file. It will return errors if
// there are errors reading the files or critical errors in the structure.
func OpenFile(idx, db string) (*File, error) {
	var d File

	df, err := os.OpenFile(idx, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("could not open index: %v", err)
	}
	defer df.Close()

	buf := make([]byte, len(FileVer))
	_, err = df.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("could not read version string: %v", err)
	} else if !bytes.Equal(buf, []byte(FileVer)) {
		return nil, fmt.Errorf("incompatible file versions: expected %#v, got %#v", FileVer, string(buf))
	}

	zr, err := zlib.NewReader(df)
	if err != nil {
		return nil, fmt.Errorf("could not decompress index: %v", err)
	}
	defer zr.Close()

	err = gob.NewDecoder(zr).Decode(&d.idx)
	if err != nil {
		return nil, fmt.Errorf("could not read index: %v", err)
	}

	debug.FreeOSMemory()

	d.df, err = os.OpenFile(db, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %v", err)
	}

	buf = make([]byte, len(FileVer))
	_, err = d.df.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("could not read version string: %v", err)
	} else if !bytes.Equal(buf, []byte(FileVer)) {
		return nil, fmt.Errorf("incompatible file versions: expected %#v, got %#v", FileVer, string(buf))
	}

	return &d, nil
}

// Verify verifies the consistency of the data structures in the database.
// WARNING: Verify takes a few seconds to run.
func (d *File) Verify() error {
	for word, cur := range d.idx {
		w, err := d.get(cur)
		if err != nil {
			return fmt.Errorf("failed: %s@%d: %v", word, cur, err)
		}
		if w.Word == "" {
			return fmt.Errorf("failed: %s@%d: empty word", word, cur)
		}
	}
	debug.FreeOSMemory()
	return nil
}

// Close closes the files associated with the dictionary file and
// clears the in-memory index. Usage of the File afterwards
// may result in a panic.
func (d *File) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.idx = nil
	return d.df.Close()
}

// HasWord implements Store.
func (d *File) HasWord(word string) bool {
	_, ok := d.idx[word]
	return ok
}

// GetWord implements Store, and will return an error if the data structure
// is invalid or the underlying files are inaccessible.
func (d *File) GetWord(word string) (*Word, bool, error) {
	cur, ok := d.idx[word]
	if !ok {
		return nil, false, nil
	}
	w, err := d.get(cur)
	return w, true, err
}

// NumWords implements Store.
func (d *File) NumWords() int {
	return len(d.idx)
}

// Lookup is a shortcut for Lookup.
func (d *File) Lookup(word string) (*Word, bool, error) {
	return Lookup(d, word)
}

// get retrieves the word at the offset in the database.
func (d *File) get(cur int64) (*Word, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.df.Seek(cur, 0)
	if err != nil {
		return nil, fmt.Errorf("could not seek to %d in db: %v", cur, err)
	}

	var n int64
	err = binary.Read(d.df, binary.LittleEndian, &n)
	if err != nil {
		return nil, fmt.Errorf("could not get gob length: %v", err)
	}

	zr, err := zlib.NewReader(io.LimitReader(d.df, n))
	if err != nil {
		return nil, fmt.Errorf("could not decompress gob: %v", err)
	}
	defer zr.Close()

	var w Word
	err = gob.NewDecoder(zr).Decode(&w)
	if err != nil {
		return nil, fmt.Errorf("could not read gob: %v", err)
	}

	return &w, nil
}
