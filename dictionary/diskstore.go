package dictionary

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/vmihailenco/msgpack/v4"
)

// FileVer is the current compatibility level of saved Files.
const FileVer = "DICT6\x00" // note: can currently handle "DICT5\x00" too

// File implements an efficient Store which is faster to initialize and uses a lot less memory (~15 MB total) than WordMap.
//
// There needs to be enough memory to store the whole index. Reading a dict is also completely thread-safe. Corrupt files
// will be detected during the read of the corrupted word (or the initialization in the case of index corruption)
// or during Verify.
//
// The dict file is stored in the following format:
//
//   + --------- + ------------ + --------------------------------------------- + ---------- + ------------------------------------------------- +
//   |           |              |  + ---- + ---------------------------- +      |            |                                                   |
//   |  FileVer  |  idx offset  |  | size | zlib compressed Word msgpack | ...  |  idx size  |  zlib compressed idx map[string][]offset msgpack  |
//   |           |              |  + =================================== +      |            |                                                   |
//   + --------- + ------------ + --------------------------------------------- + ============================================================== +
//
//   All sizes and offsets are little-endian int64. All sizes are the size of the size plus the data.
//
// The file is opened using the following steps:
//
// 1. The FileVer is read and checked. It must match exactly.
// 2. The idx offset is read.
// 3. The file is seeked to the beginning plus the idx offset.
// 4. The idx size is read.
// 5. The bytes for the idx are decompressed using zlib, and the resulting msgpack is decoded into an in-memory map[string][]int64 of the words to offsets.
//
// To read a word:
//
// 1. The offset is retrieved from the in-memory idx.
// 2. The file is seeked to the beginning plus the offset.
// 4. The size of the compressed word is read.
// 5. The bytes for the word are decompressed using zlib, and the resulting msgpack is decoded into an in-memory *Word.
//
// For more details, see the source code.
//
// It is up to the creator to ensure there aren't duplicate references to
// entries for headwords in the index. If duplicates are found, they will be
// returned as-is.
type File struct {
	idx map[string][]size
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

	// data

	var verendoff int64
	if _, err := f.Write(make([]byte, len(FileVer))); err != nil {
		return fmt.Errorf("could not write version placeholder: %v", err)
	} else if verendoff, err = f.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("could not get version end offset: %v", err)
	} else if verendoff-int64(len(FileVer)) != 0 {
		panic("bug: incorrect verendoff")
	}

	var idxoffendoff int64
	if err := size(0).Write(f); err != nil {
		return fmt.Errorf("could not write idx offset placeholder: %v", err)
	} else if idxoffendoff, err = f.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("could not get idx offset end offset: %v", err)
	} else if idxoffendoff-sizew != verendoff {
		panic("bug: incorrect idxoffendoff")
	}

	var dataendoff int64
	idx := map[string][]size{}
	if err := func() error {
		var wordendoff int64 = idxoffendoff
		var t int64 // for testing
		rev := map[*Word]size{}
		for k, ws := range wm {
			for _, w := range ws {
				if wordoff, ok := rev[w]; ok {
					idx[k] = append(idx[k], wordoff)
					continue
				}
				rev[w] = size(wordendoff)
				idx[k] = append(idx[k], rev[w])

				x, err := f.Seek(0, io.SeekCurrent) // for testing
				if err != nil {
					return fmt.Errorf("could not word start offset: %v", err)
				}

				if err = size(0).Write(f); err != nil {
					return fmt.Errorf("could not get word size placeholder: %v", err)
				} else if zw, err := zlib.NewWriterLevel(f, zlib.BestCompression); err != nil {
					return fmt.Errorf("could not compress word: %v", err)
				} else if err = msgpack.NewEncoder(zw).Encode(w); err != nil {
					return fmt.Errorf("could not encode word: %v", err)
				} else if err = zw.Close(); err != nil {
					return fmt.Errorf("could not compress word: %v", err)
				} else if wordendoff, err = f.Seek(0, io.SeekCurrent); err != nil {
					return fmt.Errorf("could not get word end offset: %v", err)
				}

				t += wordendoff - x // for testing

				if _, err := f.Seek(int64(rev[w]), io.SeekStart); err != nil {
					return fmt.Errorf("could not seek to word size placeholder: %v", err)
				} else if err := size(wordendoff - int64(rev[w])).Write(f); err != nil {
					return fmt.Errorf("could not write word size: %v", err)
				} else if wordendoff-t != idxoffendoff {
					panic("bug: incorrect wordendoff")
				}

				if _, err := f.Seek(wordendoff, io.SeekStart); err != nil {
					return fmt.Errorf("could not seek to end of current word: %v", err)
				}
			}
		}
		dataendoff = idxoffendoff + t
		return nil
	}(); err != nil {
		return fmt.Errorf("could not write data: %v", err)
	} else if x, err := f.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("could not get data end offset: %v", err)
	} else if dataendoff != x {
		panic("bug: incorrect dataendoff")
	}

	var idxendoff int64
	if err := size(0).Write(f); err != nil {
		return fmt.Errorf("could not write idx size placeholder: %v", err)
	} else if zw, err := zlib.NewWriterLevel(f, zlib.BestCompression); err != nil {
		return fmt.Errorf("could not compress idx: %v", err)
	} else if err = msgpack.NewEncoder(zw).SortMapKeys(false).UseCompactEncoding(true).Encode(idx); err != nil {
		return fmt.Errorf("could not encode idx: %v", err)
	} else if err = zw.Close(); err != nil {
		return fmt.Errorf("could not compress idx: %v", err)
	} else if idxendoff, err = f.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("could not get idx end offset: %v", err)
	} else if idxendoff-dataendoff <= sizew {
		panic("bug: incorrect idxendoff")
	}

	// offsets/sizes/version

	if _, err := f.Seek(int64(verendoff), io.SeekStart); err != nil {
		return fmt.Errorf("could not seek to idx offset placeholder: %v", err)
	} else if err = size(dataendoff).Write(f); err != nil {
		return fmt.Errorf("could not write idx offset: %v", err)
	}

	_ = dataendoff - idxoffendoff // word offsets in the data section were written earlier

	if _, err := f.Seek(dataendoff, io.SeekStart); err != nil {
		return fmt.Errorf("could not seek to idx size placeholder: %v", err)
	} else if err = size(idxendoff - dataendoff).Write(f); err != nil {
		return fmt.Errorf("could not write idx size: %v", err)
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("could not seek to version placeholder: %v", err)
	} else if _, err = f.WriteString(FileVer); err != nil {
		return fmt.Errorf("could not write version: %v", err)
	}

	// cleanup

	if err := f.Sync(); err != nil {
		return fmt.Errorf("could not write file: %v", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("could not write file: %v", err)
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

	var compat int
	buf := make([]byte, len(FileVer))
	_, err = d.df.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("could not read version string: %v", err)
	} else if bytes.Equal(buf, []byte(FileVer)) {
		compat = 6
	} else if bytes.Equal(buf, []byte("DICT5\x00")) {
		compat = 5
	} else if bytes.Equal(buf, []byte("\x00\x00\x00\x00\x00\x00")) {
		return nil, fmt.Errorf("incomplete file version: did it create successfully?")
	} else {
		return nil, fmt.Errorf("incompatible file versions: expected %#v, got %#v", FileVer, string(buf))
	}

	var idxoff, idxsize size
	if err := (&idxoff).Read(d.df); err != nil {
		return nil, fmt.Errorf("could not read idx offset: %v", err)
	}

	if err := (&idxsize).Read(io.NewSectionReader(d.df, int64(idxoff), sizew)); err != nil {
		return nil, fmt.Errorf("could not read idx size: %v", err)
	}

	zr, err := zlib.NewReader(io.NewSectionReader(d.df, int64(idxoff)+sizew, int64(idxsize)-sizew))
	if err != nil {
		return nil, fmt.Errorf("could not decompress idx: %v", err)
	}
	defer zr.Close()

	if compat >= 6 {
		if err := msgpack.NewDecoder(zr).Decode(&d.idx); err != nil {
			return nil, fmt.Errorf("could not read idx: %v", err)
		}
	} else {
		var oidx map[string]size
		if err := msgpack.NewDecoder(zr).Decode(&oidx); err != nil {
			return nil, fmt.Errorf("could not read idx: %v", err)
		}
		d.idx = make(map[string][]size, len(oidx))
		for w, o := range oidx {
			d.idx[w] = []size{o}
		}
	}

	debug.FreeOSMemory()

	return &d, nil
}

// Verify verifies the consistency of the data structures in the dict file.
// WARNING: Verify takes a few seconds to run.
func (d *File) Verify() error {
	for word, cur := range d.idx {
		for _, o := range cur {
			if w, err := d.get(o); err != nil {
				return fmt.Errorf("failed: %s#%d@%d: %v", word, o, cur, err)
			} else if w.Word == "" {
				return fmt.Errorf("failed: %s#%d@%d: empty word", word, o, cur)
			}
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
func (d *File) GetWords(word string) ([]*Word, bool, error) {
	cur, ok := d.idx[word]
	if !ok {
		return nil, false, nil
	}
	ws := make([]*Word, len(cur))
	for i, o := range cur {
		if w, err := d.get(o); err != nil {
			return nil, true, fmt.Errorf("get %s#%d@%d", word, o, i)
		} else {
			ws[i] = w
		}
	}
	return ws, true, nil
}

// GetWord is deprecated.
func (d *File) GetWord(word string) (*Word, bool, error) {
	ws, exists, err := d.GetWords(word)
	if len(ws) == 0 {
		return nil, exists, err
	}
	return ws[0], exists, err
}

// NumWords implements Store.
func (d *File) NumWords() int {
	return len(d.idx)
}

// Lookup is a shortcut for Lookup.
func (d *File) LookupWord(word string) ([]*Word, bool, error) {
	return LookupWord(d, word)
}

// Lookup is deprecated.
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
	if err := msgpack.NewDecoder(zr).Decode(&w); err != nil {
		return nil, fmt.Errorf("could not read gob: %v", err)
	}

	return &w, nil
}
