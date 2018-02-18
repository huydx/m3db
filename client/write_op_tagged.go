// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package client

import (
	"math"

	"github.com/m3db/m3db/generated/thrift/rpc"
	"github.com/m3db/m3x/ident"
	"github.com/m3db/m3x/pool"
)

var (
	// NB(bl): use an invalid shardID for the zerod op
	writeOpTaggedZeroed = writeOpTagged{shardID: math.MaxUint32}
)

type writeOpTagged struct {
	namespace    ident.ID
	shardID      uint32
	request      rpc.WriteTaggedBatchRawRequestElement
	datapoint    rpc.Datapoint
	completionFn completionFn
	pool         *writeOpTaggedPool
}

func (w *writeOpTagged) reset() {
	*w = writeOpTaggedZeroed
	w.request.Datapoint = &w.datapoint
}

func (w *writeOpTagged) Close() {
	p := w.pool
	w.reset()
	if p != nil {
		p.Put(w)
	}
}

func (w *writeOpTagged) Size() int {
	// Writes always represent a single write
	return 1
}

func (w *writeOpTagged) CompletionFn() completionFn {
	return w.completionFn
}

func (w *writeOpTagged) SetCompletionFn(fn completionFn) {
	w.completionFn = fn
}

func (w *writeOpTagged) ShardID() uint32 {
	return w.shardID
}

type writeOpTaggedPool struct {
	pool pool.ObjectPool
}

func newWriteOpTaggedPool(opts pool.ObjectPoolOptions) *writeOpTaggedPool {
	p := pool.NewObjectPool(opts)
	return &writeOpTaggedPool{pool: p}
}

func (p *writeOpTaggedPool) Init() {
	p.pool.Init(func() interface{} {
		w := &writeOpTagged{}
		w.reset()
		return w
	})
}

func (p *writeOpTaggedPool) Get() *writeOpTagged {
	w := p.pool.Get().(*writeOpTagged)
	w.pool = p
	return w
}

func (p *writeOpTaggedPool) Put(w *writeOpTagged) {
	w.reset()
	p.pool.Put(w)
}
