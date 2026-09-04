package bytecode

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/version"
	"github.com/dchest/siphash"
)

// The bytecode cache in __engcache__ lets a program that has not changed skip
// parsing. Whether a cached file may be used is the whole question, and it used
// to be answered by comparing modification times: valid if the cache file was
// not older than the source.
//
// That is wrong in both directions. A checkout, a restore from a backup, an
// archive extraction or an editor that preserves timestamps can leave a source
// file *older* than a cache built from different text, and the stale cache is
// then used silently — the program you edited is not the program that runs.
// Meanwhile `touch` on an unchanged file throws away a perfectly good cache.
// And nothing recorded which build wrote the file, so a cache written by an
// older compiler was decoded by a newer one whose node numbering may since
// have changed, which is a decode error at best and a wrong AST at worst.
//
// The header quotes PEP 552, which exists specifically to replace timestamp
// invalidation with a hash of the source. This does that: a cache file carries
// a stamp over the compiler version, the format version and the source bytes,
// and is used only when that stamp still matches.

// CacheDir is the directory cached bytecode is written to.
const CacheDir = "__engcache__"

// cacheMagic marks a cache envelope. It deliberately differs from MagicBytes:
// a cache file is not a compiled artefact, and handing one to a tool that
// expects a .101 program should say so rather than half-work.
var cacheMagic = []byte{0x10, 0x1E, 0x43, 0x41}

// cacheEnvelopeVersion is the version of the envelope itself, so its shape can
// change without being mistaken for a stamp mismatch.
const cacheEnvelopeVersion uint8 = 1

// cacheHeaderSize is the magic, the envelope version and the 8-byte stamp.
const cacheHeaderSize = 4 + 1 + 8

// Cache configuration
const (
	// CacheHashBytes is the number of bytes of SipHash used for cache
	// filenames and stamps. SipHash produces an 8-byte (64-bit) hash, which
	// resists collision well enough here and is far faster than a
	// cryptographic hash. Using the full 8 bytes, as PEP 552 recommends.
	CacheHashBytes = 8
)

// Fixed keys, so a stamp is reproducible across runs and machines. This is a
// cache key, not a security boundary: a hostile source file can already do
// anything the program it describes can do.
const (
	cacheKey0 = uint64(0x0706050403020100)
	cacheKey1 = uint64(0x0f0e0d0c0b0a0908)
)

// GetCachePath returns the cache file path for a given source file.
// For example: "examples/math_library.abc" -> "__engcache__/39ccbccfa9db97df_math_library.abc.101"
//
// The name identifies the *source*, so a file has one cache entry that is
// replaced when it changes rather than accumulating one per revision. Whether
// that entry is current is what the stamp answers.
func GetCachePath(sourcePath string) string {
	hash := siphash.Hash(cacheKey0, cacheKey1, []byte(sourcePath))
	return filepath.Join(CacheDir, fmt.Sprintf("%016x_%s.101", hash, filepath.Base(sourcePath)))
}

// CacheStamp is the value a cache entry is keyed on: the compiler version, the
// bytecode format version and the source text.
//
// Including the compiler version means a cache written by a different build is
// never read, which is what makes it safe to change the encoding at all.
func CacheStamp(source []byte) uint64 {
	var buf bytes.Buffer
	buf.WriteString(version.Version)
	buf.WriteByte(0)
	buf.WriteByte(FormatVersion)
	buf.WriteByte(0)
	buf.Write(source)
	return siphash.Hash(cacheKey0, cacheKey1, buf.Bytes())
}

// StampOfSource reads a source file and returns its stamp.
func StampOfSource(sourcePath string) (uint64, error) {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return 0, err
	}
	return CacheStamp(source), nil
}

// IsCacheValid reports whether the cached bytecode was built from exactly this
// source text by this build of the compiler.
func IsCacheValid(sourcePath, cachePath string) bool {
	want, err := StampOfSource(sourcePath)
	if err != nil {
		return false
	}
	got, err := stampOfCache(cachePath)
	if err != nil {
		return false
	}
	return got == want
}

// stampOfCache reads just the header of a cache file.
func stampOfCache(cachePath string) (uint64, error) {
	file, err := os.Open(cachePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	header := make([]byte, cacheHeaderSize)
	if _, err := io.ReadFull(file, header); err != nil {
		return 0, fmt.Errorf("cache file is truncated: %w", err)
	}
	if !bytes.Equal(header[:4], cacheMagic) {
		return 0, fmt.Errorf("not a bytecode cache file")
	}
	if header[4] != cacheEnvelopeVersion {
		return 0, fmt.Errorf("cache envelope version %d, expected %d", header[4], cacheEnvelopeVersion)
	}
	return binary.LittleEndian.Uint64(header[5:]), nil
}

// WriteBytecodeCache writes bytecode to the cache directory, stamped with the
// source it was built from. Creates the cache directory if it doesn't exist.
func WriteBytecodeCache(cachePath string, stamp uint64, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	var buf bytes.Buffer
	buf.Grow(cacheHeaderSize + len(payload))
	buf.Write(cacheMagic)
	buf.WriteByte(cacheEnvelopeVersion)
	var stampBytes [8]byte
	binary.LittleEndian.PutUint64(stampBytes[:], stamp)
	buf.Write(stampBytes[:])
	buf.Write(payload)

	if err := os.WriteFile(cachePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	return nil
}

// ReadBytecodeCache reads a cache file, returning the bytecode it holds and
// the stamp it was written with.
func ReadBytecodeCache(cachePath string) ([]byte, uint64, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read cache file: %w", err)
	}
	if len(data) < cacheHeaderSize {
		return nil, 0, fmt.Errorf("cache file is truncated")
	}
	if !bytes.Equal(data[:4], cacheMagic) {
		return nil, 0, fmt.Errorf("not a bytecode cache file")
	}
	if data[4] != cacheEnvelopeVersion {
		return nil, 0, fmt.Errorf("cache envelope version %d, expected %d", data[4], cacheEnvelopeVersion)
	}
	return data[cacheHeaderSize:], binary.LittleEndian.Uint64(data[5:cacheHeaderSize]), nil
}

// LoadFromCache decodes a program from its cache entry, if that entry was
// built from exactly this source by this build.
//
// The second result is the path the program was read from, for a tool that
// wants to say where it came from; the third reports whether anything was
// found at all.
func LoadFromCache(sourcePath string) (*ast.Program, string, bool) {
	cachePath := GetCachePath(sourcePath)
	stamp, err := StampOfSource(sourcePath)
	if err != nil {
		return nil, "", false
	}
	payload, cached, err := ReadBytecodeCache(cachePath)
	if err != nil || cached != stamp {
		return nil, "", false
	}
	program, err := NewDecoder(payload).Decode()
	if err != nil {
		return nil, "", false
	}
	return program, cachePath, true
}

// LoadCachedOrParse loads a program from the cache when the cache was built
// from this exact source, and otherwise parses and caches it.
//
// The second result reports whether the cache was used.
func LoadCachedOrParse(sourcePath string, parseFunc func(string) (*ast.Program, error)) (*ast.Program, bool, error) {
	cachePath := GetCachePath(sourcePath)

	stamp, stampErr := StampOfSource(sourcePath)
	if stampErr == nil {
		payload, cached, err := ReadBytecodeCache(cachePath)
		if err == nil && cached == stamp {
			program, err := NewDecoder(payload).Decode()
			if err == nil {
				return program, true, nil
			}
			// The stamp matched but the contents did not decode, so the file
			// is damaged rather than stale. Remove it: leaving it in place
			// meant every later run re-read and re-rejected the same bytes.
			_ = os.Remove(cachePath)
		}
	}

	program, err := parseFunc(sourcePath)
	if err != nil {
		return nil, false, err
	}

	// Caching is an optimisation, so a failure to write one is not a failure
	// to run the program: a read-only directory or a full disk should not stop
	// execution. A stamp we could not compute is a different matter — without
	// it the entry could never be validated, so nothing is written.
	if stampErr == nil {
		if data, err := NewEncoder().Encode(program); err == nil {
			_ = WriteBytecodeCache(cachePath, stamp, data)
		}
	}

	return program, false, nil
}
