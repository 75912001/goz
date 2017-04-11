package zutility

import (
	"crypto/md5"
	"encoding/hex"
	"hash/fnv"
)

////////////////////////////////////////////////////////////////////////////////
//md5
func GenMd5(s *string) (value string) {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(*s))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func HASH(s *string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(*s))
	return h.Sum32()
}
