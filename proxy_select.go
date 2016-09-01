package main

import (
	"github.com/tarantool/go-tarantool"
)

func (self *ProxyConnection) executeRequestSelect(requestType uint32, requestId uint32,
	reader IprotoReader) (flags uint32, response *tarantool.Response, err error) {
	//|--------------- header ----------------|---------------request_body ---------------------...|
	// <request_type><body_length><request_id> <space_no><index_no><offset><limit><count><tuple>+
	var (
		spaceNo     uint32
		indexNo     uint32
		offset      uint32
		limit       uint32
		count       uint32
		cardinality uint32
		args        []interface{}
		param       interface{}
	)

	unpackUint32(reader, &spaceNo)
	unpackUint32(reader, &indexNo)
	unpackUint32(reader, &offset)
	unpackUint32(reader, &limit)
	unpackUint32(reader, &count)

	space, err := self.schema.GetSpaceInfo(spaceNo)
	if err != nil {
		return
	}

	indexName, err := space.GetIndexName(indexNo)
	if err != nil {
		return
	}

	indexDefs, err := space.GetIndexDefs(indexNo)
	if err != nil {
		return
	}

	// далее лежат упакованные iproto tuple-ы
	for i := uint32(0); i < count; i += 1 {
		err = unpackUint32(reader, &cardinality)
		if err != nil {
			return
		}

		for fieldNo := uint32(0); fieldNo < cardinality; fieldNo += 1 {
			param, err = self.unpackFieldByDefs(reader, requestType, fieldNo, indexDefs[fieldNo])
			if err != nil {
				return
			}
			args = append(args, param)
		} //end for
	} //end for

	//sharding for key0
	tnt16 := self.getTnt16(args[0])

	response, err = tnt16.Select(space.name, indexName, offset, limit, tarantool.IterEq, args)
	return
}