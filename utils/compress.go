package cherryUtils

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
)

type compress struct {
}

func (c *compress) DeflateData(data []byte) ([]byte, error) {
	var bb bytes.Buffer
	z := zlib.NewWriter(&bb)
	_, err := z.Write(data)
	if err != nil {
		return nil, err
	}
	z.Close()
	return bb.Bytes(), nil
}

func (c *compress) InflateData(data []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	zr.Close()
	return ioutil.ReadAll(zr)
}

func (c *compress) IsCompressed(data []byte) bool {
	return len(data) > 2 &&
		(
		//zlib
		(data[0] == 0x78 &&
			(data[1] == 0x9C ||
				data[1] == 0x01 ||
				data[1] == 0xDA ||
				data[1] == 0x5E)) ||
			// gzip
			(data[0] == 0x1F && data[1] == 0x8B))
}
