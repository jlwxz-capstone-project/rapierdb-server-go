package storage

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

type LoadedDoc struct {
	DocID string
	Doc   *loro.LoroDoc
}

type DocsCache struct {
	docs map[string]*LoadedDoc
}

func NewEmptyDocsCache() *DocsCache {
	return &DocsCache{
		docs: make(map[string]*LoadedDoc),
	}
}

func (c *DocsCache) Get(key []byte) *loro.LoroDoc {
	doc, ok := c.docs[string(key)]
	if !ok {
		return nil
	}
	return doc.Doc
}

func (c *DocsCache) Set(key []byte, docID string, doc *loro.LoroDoc) {
	c.docs[string(key)] = &LoadedDoc{DocID: docID, Doc: doc}
}

func (c *DocsCache) Delete(key []byte) {
	delete(c.docs, string(key))
}

func (c *DocsCache) Clear() {
	c.docs = make(map[string]*LoadedDoc)
}

func (c *DocsCache) DocCount() int {
	return len(c.docs)
}
