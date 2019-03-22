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
)

// FileVer is the current compatibility level of saved Files.
const FileVer = "DICT4\x00"

// File implements an efficient Store which is faster to initialize and uses a lot less memory (~15 MB total) than WordMap.
//
// There needs to be enough memory to store the whole index. Reading a dict is also completely thread-safe. Corrupt files
// will be detected during the read of the corrupted word (or the initialization in the case of index corruption)
// or during Verify.
//
// The dict file is stored in the following format:
//
//   +-----------+--------------+-------------------------------------------+------------+---------------------------------------------+
//   |           |              |  +------+--------------------------+      |            |                                             |
//   |  FileVer  |  idx offset  |  | size | zlib compressed Word gob | ...  |  idx size  |  zlib compressed idx map[string]offset gob  |
//   |           |              |  +------+--------------------------+      |            |                                             |
//   +-----------+--------------+-------------------------------------------+---- -------+---------------------------------------------+
//
//   All sizes and offsets are little-endian int64.
//
// The file is opened using the following steps:
//
// 1. The FileVer is read and checked. It must match exactly.
// 2. The idx offset is read.
// 3. The file is seeked to the beginning plus the idx offset.
// 4. The idx size is read.
// 5. The bytes for the idx are decompressed using zlib, and the resulting gob is decoded into an in-memory map[string]int64 of the words to offsets.
//
// To read a word:
//
// 1. The offset is retrieved from the in-memory idx.
// 2. The file is seeked to the beginning plus the offset.
// 4. The size of the compressed word is read.
// 5. The bytes for the word are decompressed using zlib, and the resulting gob is decoded into an in-memory *Word.
//
// To create the file:
// 1. The FileVer is written.
// 2. A placeholder for the idx offset is written.
// 3. The dictionary map is looped over:
//    a. If the referenced Word has already been written, it is skipped.
//    b. The Word is encoded as a gob, compressed and written to a temporary buf.
//    c. The size is written.
//    d. The buf is written and reset.
// 5. The current offset is stored for the idx offset.
// 4. A placeholder for the idx size is written.
// 5. The idx is encoded as a gob, compressed, and written.
// 6. The idx size is calculated by subtracting the idx offset and the size of the idx size from the current offset.
// 7. The file is seeked to the idx offset and the idx size is written.
// 8. The file is seeked to the beginning plus the length of FileVer.
// 9. The idx offset is written.
//
type File struct {
	idx map[string]size
	df  interface {
		io.Reader
		io.Closer
		io.ReaderAt
	}
}

type size int64

var sizew = int64(binary.Size(size(0)))

// Write writes the size to a writer. It is guaranteed to write sizew
// bytes on success.
func (s size) Write(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, int64(s))
}

// Read reads a size from a reader. It is guaranteed to read sizew
// bytes on success. It should be called like (&s).Read(r).
func (s *size) Read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, s)
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

	if _, err := f.WriteString(FileVer); err != nil {
		return fmt.Errorf("could not write version string: %v", err)
	}

	if err := size(0).Write(f); err != nil {
		return fmt.Errorf("could not write idx offset placeholder: %v", err)
	}

	idx := map[string]size{}
	revidx := map[*Word]size{}
	buf := new(bytes.Buffer)
	for k, v := range wm {
		if cur, ok := revidx[v]; ok {
			idx[k] = cur
			continue
		}

		buf.Reset()

		cur, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return fmt.Errorf("could not get word offset: %v", err)
		}

		zw, err := zlib.NewWriterLevel(buf, zlib.BestCompression)
		if err != nil {
			return fmt.Errorf("could not compress word: %v", err)
		}

		if err := gob.NewEncoder(zw).Encode(v); err != nil {
			return fmt.Errorf("could not encode word: %v", err)
		}

		zw.Close()

		if err := size(buf.Len()).Write(f); err != nil {
			return fmt.Errorf("could not write word size: %v", err)
		}

		if _, err := f.Write(buf.Bytes()); err != nil {
			return fmt.Errorf("could not write word: %v", err)
		}

		idx[k] = size(cur)
		revidx[v] = size(cur)
	}

	idxoff, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("could not get idx offset: %v", err)
	}

	if err := size(0).Write(f); err != nil {
		return fmt.Errorf("could not write idx size placeholder: %v", err)
	}

	zw, err := zlib.NewWriterLevel(f, zlib.BestCompression)
	if err != nil {
		return fmt.Errorf("could not compress idx: %v", err)
	}

	if err := gob.NewEncoder(zw).Encode(idx); err != nil {
		return fmt.Errorf("could not encode idx: %v", err)
	}

	zw.Close() // must be called here before the seek

	idxendoff, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("could not get current offset: %v", err)
	}

	if _, err := f.Seek(idxoff, io.SeekStart); err != nil {
		return fmt.Errorf("could not seek to idx size placeholder: %v", err)
	}

	if err := size(idxendoff - idxoff).Write(f); err != nil {
		return fmt.Errorf("could not write idx size: %v", err)
	}

	if _, err := f.Seek(int64(len([]byte(FileVer))), io.SeekStart); err != nil {
		return fmt.Errorf("could not seek to idx offset placeholder: %v", err)
	}

	if err := size(idxoff).Write(f); err != nil {
		return fmt.Errorf("could not write idx offset: %v", err)
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

	var idxoff, idxsize size
	if err := (&idxoff).Read(d.df); err != nil {
		return nil, fmt.Errorf("could not read idx offset: %v", err)
	}

	if err := (&idxsize).Read(io.NewSectionReader(d.df, int64(idxoff), sizew)); err != nil {
		return nil, fmt.Errorf("could not read idx size: %v", err)
	}

	zr, err := zlib.NewReader(io.NewSectionReader(d.df, int64(idxoff)+sizew, int64(idxsize)))
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
		if w, err := d.get(cur); err != nil {
			return fmt.Errorf("failed: %s@%d: %v", word, cur, err)
		} else if w.Word == "" {
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
func (d *File) get(cur size) (*Word, error) {
	var n int64
	if err := binary.Read(io.NewSectionReader(d.df, int64(cur), int64(binary.Size(n))), binary.LittleEndian, &n); err != nil {
		return nil, fmt.Errorf("could not get gob length: %v", err)
	}

	zr, err := zlib.NewReader(io.NewSectionReader(d.df, int64(cur)+sizew, n))
	if err != nil {
		return nil, fmt.Errorf("could not decompress gob: %v", err)
	}
	defer zr.Close()

	var w Word
	if err := gob.NewDecoder(zr).Decode(&w); err != nil {
		return nil, fmt.Errorf("could not read gob: %v", err)
	}

	return &w, nil
}
