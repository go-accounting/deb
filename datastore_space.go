// +build appengine

package deb

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"runtime"
	"time"

	"appengine"
	"appengine/datastore"
)

type datastoreSpace struct{}

type blockWrapper struct {
	D    []byte
	AsOf time.Time
}

type keyWrapper struct {
	key  *datastore.Key
	asOf time.Time
}

type errorWithStackTrace struct {
	err   error
	stack [4096]byte
}

func newErrorWithStackTrace(err error) errorWithStackTrace {
	ewst := errorWithStackTrace{err: err}
	runtime.Stack(ewst.stack[:], false)
	return ewst
}

func (e errorWithStackTrace) Error() string {
	return fmt.Sprintf("%q\n%s\n", e.err, e.stack[:])
}

func NewDatastoreSpace(ctx appengine.Context, key *datastore.Key) (Space, *datastore.Key, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("ctx is nil")
	}
	if key == nil {
		key = datastore.NewIncompleteKey(ctx, "space", nil)
		var err error
		if key, err = datastore.Put(ctx, key, &datastoreSpace{}); err != nil {
			return nil, nil, err
		}
	}
	var ls *largeSpace
	errc := make(chan error, 1)
	in := func() chan *dataBlock {
		c := make(chan *dataBlock)
		go func() {
			var err error
			defer func() {
				close(c)
				errc <- err
			}()
			q := datastore.NewQuery("data_block").Ancestor(key).Order("AsOf")
			t := q.Run(ctx)
			for {
				bw := blockWrapper{}
				var k *datastore.Key
				k, err = t.Next(&bw)
				if err == datastore.Done {
					err = nil
					break
				}
				if err != nil {
					err = newErrorWithStackTrace(err)
					break
				}
				buf := bytes.NewBuffer(bw.D)
				dec := gob.NewDecoder(buf)
				block := ls.newDataBlock()
				if err = dec.Decode(block); err != nil {
					err = newErrorWithStackTrace(err)
					break
				}
				block.key = keyWrapper{k, bw.AsOf}
				c <- block
			}
		}()
		return c
	}
	out := make(chan []*dataBlock)
	go func() {
		for blocks := range out {
			errc <- datastore.RunInTransaction(ctx, func(tc appengine.Context) error {
				var err error
				keys := make([]*datastore.Key, 0, len(blocks))
				asOfs := make([]time.Time, 0, len(blocks))
				storedAsOfs := make([]*struct{ AsOf time.Time }, 0, len(blocks))
				for _, block := range blocks {
					if block.key == nil || block.key.(keyWrapper).key == nil {
						block.key = keyWrapper{
							datastore.NewIncompleteKey(tc, "data_block", key), time.Now()}
					} else {
						keys = append(keys, block.key.(keyWrapper).key)
						asOfs = append(asOfs, block.key.(keyWrapper).asOf)
					}
				}
				storedAsOfs = storedAsOfs[0:len(keys)]
				if err = datastore.GetMulti(tc, keys, storedAsOfs); err != nil {
					if merr, ok := err.(appengine.MultiError); ok {
						found := false
						for _, err := range merr {
							if _, ok := err.(*datastore.ErrFieldMismatch); !ok {
								found = true
								break
							}
						}
						if found {
							return merr
						}
						err = nil
					} else {
						return newErrorWithStackTrace(err)
					}
				}
				for i, _ := range keys {
					if asOfs[i] != storedAsOfs[i].AsOf {
						return fmt.Errorf("Concurrent modification")
					}
				}
				keys = make([]*datastore.Key, len(blocks))
				bws := make([]*blockWrapper, len(blocks))
				for i, block := range blocks {
					var buf bytes.Buffer
					enc := gob.NewEncoder(&buf)
					if err = enc.Encode(block); err != nil {
						err = newErrorWithStackTrace(err)
						break
					}
					keys[i] = block.key.(keyWrapper).key
					bws[i] = &blockWrapper{buf.Bytes(), time.Now()}
				}
				if _, err = datastore.PutMulti(tc, keys, bws); err != nil {
					return newErrorWithStackTrace(err)
				}
				return nil
			}, nil)
		}
	}()
	ls = newLargeSpace(1014*1024, in, out, errc)
	return ls, key, nil
}

func CopySpaceToDatastore(ctx appengine.Context, key *datastore.Key, space Space) error {
	if ls, ok := space.(*largeSpace); !ok {
		return fmt.Errorf("Not a largeSpace")
	} else {
		ctx.Infof("Starting copying space")
		var err error
		for b := range ls.in() {
			if err != nil {
				continue
			}
			bk := datastore.NewIncompleteKey(ctx, "data_block", key)
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			if err = enc.Encode(b); err != nil {
				continue
			}
			if _, err := datastore.Put(ctx, bk,
				&blockWrapper{buf.Bytes(), time.Now()}); err != nil {
				return err
			}
		}
		if err != nil {
			<-ls.errc
			return err
		}
		return <-ls.errc
	}
}
