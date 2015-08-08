// +build appengine

package deb

import (
	"fmt"
	"time"

	"appengine"
	"appengine/datastore"
)

type datastoreSpace struct{}

type blockWrapper struct {
	Block dataBlock
	AsOf  time.Time
}

type keyWrapper struct {
	key  *datastore.Key
	asOf time.Time
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
			defer close(c)
			defer func() { errc <- err }()
			q := datastore.NewQuery("data_block").Ancestor(key)
			t := q.Run(ctx)
			for {
				bw := blockWrapper{*ls.newDataBlock(), time.Now()}
				mdSize := cap(bw.Block.MD)
				var k *datastore.Key
				k, err = t.Next(&bw)
				newMD := make([]byte, len(bw.Block.MD), mdSize)
				copy(newMD, bw.Block.MD)
				bw.Block.MD = newMD
				if err == datastore.Done {
					err = nil
					break
				}
				if err != nil {
					break
				}
				bw.Block.key = keyWrapper{k, bw.AsOf}
				c <- &bw.Block
			}
		}()
		return c
	}
	out := make(chan *dataBlock)
	go func() {
		for block := range out {
			bw := blockWrapper{*block, time.Now()}
			if block.key == nil || block.key.(keyWrapper).key == nil {
				block.key = keyWrapper{datastore.NewIncompleteKey(ctx, "data_block", key),
					time.Now()}
			} else {
				var bw2 blockWrapper
				if err := datastore.Get(ctx, block.key.(keyWrapper).key, &bw2); err != nil {
					errc <- err
					continue
				}
				if bw2.AsOf != block.key.(keyWrapper).asOf {
					errc <- fmt.Errorf("Concurrent modification")
					continue
				}
			}
			_, err := datastore.Put(ctx, block.key.(keyWrapper).key, &bw)
			errc <- err
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
		for b := range ls.in() {
			bk := datastore.NewIncompleteKey(ctx, "data_block", key)
			if _, err := datastore.Put(ctx, bk, &blockWrapper{*b, time.Now()}); err != nil {
				return err
			}
		}
		return <-ls.errc
	}
}
