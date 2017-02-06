package filetype

import (
	"errors"
	"gopkg.in/h2non/filetype.v0/matchers"
	"gopkg.in/h2non/filetype.v0/types"
)

// Map of supported types
var Types = types.Types

// Create and register a new type
var NewType = types.NewType

// Default unknown file type
var Unknown = types.Unknown

// Predefined errors
var EmptyBufferErr = errors.New("Empty buffer")
var UnknownBufferErr = errors.New("Unknown buffer type")

// Register a new file type
func AddType(ext, mime string) types.Type {
	return types.NewType(ext, mime)
}

// Checks if a given buffer matches with the given file type extension
func Is(buf []byte, ext string) bool {
	kind, ok := types.Types[ext]
	if ok {
		return IsType(buf, kind)
	}
	return false
}

// Semantic alias to Is()
func IsExtension(buf []byte, ext string) bool {
	return Is(buf, ext)
}

// Checks if a given buffer matches with the given file type
func IsType(buf []byte, kind types.Type) bool {
	matcher := matchers.Matchers[kind]
	if matcher == nil {
		return false
	}
	return matcher(buf) != types.Unknown
}

// Checks if a given buffer matches with the given MIME type
func IsMIME(buf []byte, mime string) bool {
	for _, kind := range types.Types {
		if kind.MIME.Value == mime {
			matcher := matchers.Matchers[kind]
			return matcher(buf) != types.Unknown
		}
	}
	return false
}

// Check if a given file extension is supported
func IsSupported(ext string) bool {
	for name, _ := range Types {
		if name == ext {
			return true
		}
	}
	return false
}

// Check if a given MIME type is supported
func IsMIMESupported(mime string) bool {
	for _, m := range Types {
		if m.MIME.Value == mime {
			return true
		}
	}
	return false
}

// Retrieve a Type by file extension
func GetType(ext string) types.Type {
	return types.Get(ext)
}
