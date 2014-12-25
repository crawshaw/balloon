// Experiment in reading a Java keystore file (.jks), used for android signing.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

func main() {
	b, err := ioutil.ReadFile("/Users/crawshaw/.android/debug.keystore")
	if err != nil {
		log.Fatal(err)
	}
	jks, err := parseJKS(b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(jks)
}

type Cert struct {
	Type string
	Data []byte
}

type Key struct {
	Alias string
	Time  time.Time
	Key   []byte
	Cert  []Cert
}

type JKS []Key

func (jks JKS) String() string {
	return fmt.Sprintf("JKS%v", ([]Key)(jks))
}

func parseJKS(b []byte) (jks JKS, err error) {
	read := func(i int) []byte {
		ret := b[:i]
		b = b[i:]
		return ret
	}
	readModifiedUTF8 := func() string {
		l := uint16(b[0])<<8 | uint16(b[1])
		b = b[2:]
		return string(read(int(l))) // TODO: adjust NULLs
	}
	readTime := func() time.Time {
		// Milliseconds since Unix epoch.
		ret := int64(uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 | uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7]))
		b = b[8:]
		return time.Unix(ret/1000, int64(ret)%1000*1e6)
	}
	u32 := func() uint32 {
		ret := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
		b = b[4:]
		return ret
	}
	i32 := func() int32 { return int32(u32()) }

	// JavaKeyStore format is described (in loose Java-esque terms) in the
	// comments sun/security/provider/JavaKeyStore.java.
	if len(b) < 12 {
		return nil, fmt.Errorf("jks: too short for header, len=%d", len(b))
	}
	const magic = 0xfeedfeed
	if header := u32(); header != magic {
		return nil, fmt.Errorf("jks: bad magic %x, want %x", header, magic)
	}
	if version := i32(); version != 2 {
		return nil, fmt.Errorf("jks: version=%d, want 2", version)
	}
	for i, count := 0, int(i32()); i < count; i++ {
		tag := i32()
		alias := readModifiedUTF8()

		t := readTime()

		key := Key{
			Alias: alias,
			Time:  t,
		}

		certCount := 0
		switch tag {
		case 1:
			keyLen := int(i32())
			key.Key = read(keyLen)
			certCount = int(i32())
		case 2:
			certCount = 1
		default:
			return nil, fmt.Errorf("jks: unknown tag %d", tag)
		}

		for cert := 0; cert < certCount; cert++ {
			typ := readModifiedUTF8()
			certLen := int(i32())
			data := read(certLen)
			key.Cert = append(key.Cert, Cert{
				Type: typ,
				Data: data,
			})
		}
		jks = append(jks, key)
	}
	return jks, nil
}
