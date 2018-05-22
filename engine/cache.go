package engine

import (
	"crypto/md5"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/diskv"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/stoic-cli/stoic-cli-core"
)

func (e *engine) Cache() stoic.Cache {
	if e.cache == nil {
		cacheDir := filepath.Join(e.Root(), "cache")
		tempDir := filepath.Join(e.Root(), "temp")

		e.cache = &diskvCache{
			diskv: diskv.New(diskv.Options{
				BasePath: cacheDir,
				AdvancedTransform: func(k string) *diskv.PathKey {
					path := strings.Split(k, "/")
					filename := fmt.Sprintf("%x", md5.Sum([]byte(k)))
					return &diskv.PathKey{
						Path:     path,
						FileName: filename,
					}
				},
				InverseTransform: func(k *diskv.PathKey) string {
					return strings.Join(k.Path, "/")
				},
				CacheSizeMax: 512 * 1024 * 1024,
				TempDir:      tempDir,
			}),
		}
	}
	return e.cache
}

type diskvCache struct {
	diskv *diskv.Diskv
}

func (dvc *diskvCache) Put(key string, r io.Reader) error {
	return dvc.diskv.WriteStream(key, r, false)
}

func (dvc *diskvCache) Get(key string) io.ReadCloser {
	reader, err := dvc.diskv.ReadStream(key, false)
	if err != nil {
		jww.WARN.Printf("error loading %v from cache: %v", key, err)
		return nil
	}
	return reader
}
