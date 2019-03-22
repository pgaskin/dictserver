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
const FileVer = "DICT3\x00"

// File implements an efficient Store which is faster to initialize and uses a lot less memory (~15 MB total) than WordMap.
//
// There needs to be enough memory to store the whole index. There can be multiple concurrent readers of a single file, but each reader will read
// one word at a time. Corrupt files will be detected during the read of the corrupted word (or the initialization in the case of index corruption)
// or during Verify.
//
// The dict file is stored in the following format:
//
//   +-----------+----------------------+---------------------------------------------------+---------------------------------------------------+
//   |           |                      |  +--------------+--------------------------+      |                                                   |
//   |  FileVer  |  idx offset (int64)  |  | size (int64) | zlib compressed Word gob | ...  |  zlib compressed idx map[string]int64 offset gob  |
//   |           |                      |  +--------------+--------------------------+      |                                                   |
//   +-----------+----------------------+---------------------------------------------------+---------------------------------------------------+
//
// The file is opened using the following steps:
//
// 1. The FileVer is read and checked. It must match exactly.
// 2. The next int64 is read into a variable for idx offset.
// 3. The file is seeked to the beginning plus the idx offset.
// 4. The bytes for the idx are decompressed using zlib, and the resulting gob is decoded into an in-memory map[string]int64 of the words to offsets.
//
// To read a word:
//
// 1. The offset is retrieved from the in-memory idx.
// 2. The file is seeked to the beginning plus the offset.
// 4. The int64 is read into a variable for the size of the compressed definition.
// 5. The bytes for the definition are decompressed using zlib, and the resulting gob is decoded into an in-memory *Word.
//
// To create the file:
// 1. The FileVer is written.
// 2. A placeholder for the idx offset is written.
// 3. The dictionary map is looped over:
//    a. If the referenced Word has already been written, it is skipped.
//    b. The Word is encoded as a gob, compressed and written to a temporary buf.
//    c. The size is written.
//    d. The buf is written and reset.
// 4. The current offset is stored and the idx is encoded as a gob, compressed, and written.
// 5. The file is seeked to the beginning plus the length of FileVer.
// 6. The idx offset is written.
//
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
func CreateFile(wm WordMap, dictfile string) error {
	f, err := os.Create(dictfile)
	if err != nil {
		return fmt.Errorf("could not create db: %v", err)
	}
	defer f.Sync()
	defer f.Close()

	_, err = f.WriteString(FileVer)
	if err != nil {
		return fmt.Errorf("could not write version string: %v", err)
	}

	err = binary.Write(f, binary.LittleEndian, int64(0))
	if err != nil {
		return fmt.Errorf("could not write idx offset placeholder: %v", err)
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

		cur, err := f.Seek(0, io.SeekCurrent)
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

	idxoff, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("could not get idx offset: %v", err)
	}

	zw, err := zlib.NewWriterLevel(f, zlib.BestCompression)
	if err != nil {
		return fmt.Errorf("could not compress idx: %v", err)
	}

	err = gob.NewEncoder(zw).Encode(idx)
	if err != nil {
		return fmt.Errorf("could not encode idx: %v", err)
	}

	zw.Close() // must be called here before the seek

	_, err = f.Seek(int64(len([]byte(FileVer))), io.SeekStart)
	if err != nil {
		return fmt.Errorf("could not seek to idx offset placeholder: %v", err)
	}

	err = binary.Write(f, binary.LittleEndian, int64(idxoff))
	if err != nil {
		return fmt.Errorf("could not write idx size: %v", err)
	}

	return nil
}

// OpenFile opens a dictionary file. It will return errors if
// there are errors reading the files or critical errors in the structure.
func OpenFile(dictfile string) (*File, error) {
	var d File
	var err error

	d.df, err = os.OpenFile(dictfile, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %v", err)
	}

	buf := make([]byte, len(FileVer))
	_, err = d.df.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("could not read version string: %v", err)
	} else if !bytes.Equal(buf, []byte(FileVer)) {
		return nil, fmt.Errorf("incompatible file versions: expected %#v, got %#v", FileVer, string(buf))
	}

	var v int64
	err = binary.Read(d.df, binary.LittleEndian, &v)
	if err != nil {
		return nil, fmt.Errorf("could not read idx offset: %v", err)
	}

	_, err = d.df.Seek(v, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("could not seek to idx: %v", err)
	}

	zr, err := zlib.NewReader(d.df)
	if err != nil {
		return nil, fmt.Errorf("could not decompress idx: %v", err)
	}
	defer zr.Close()

	err = gob.NewDecoder(zr).Decode(&d.idx)
	if err != nil {
		return nil, fmt.Errorf("could not read idx: %v", err)
	}

	debug.FreeOSMemory()

	return &d, nil
}

// Verify verifies the consistency of the data structures in the dict file.
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

// get retrieves the word at the offset in the dict file.
func (d *File) get(cur int64) (*Word, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.df.Seek(cur, io.SeekStart)
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
