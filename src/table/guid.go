/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/11/13
 * Time: 10:22 AM
 * To change this template use File | Settings | File Templates.
 */
package table

// relevant bits from https://github.com/abneptis/GoUUID/blob/master/uuid.go

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type UUID [16]byte

// create a new uuid v4
func NewUUID() *UUID {
        u := &UUID{}
        _, err := rand.Read(u[:16])
        if err != nil {
                panic(err)
        }

        u[8] = (u[8] | 0x80) & 0xBf
        u[6] = (u[6] | 0x40) & 0x4f
        return u
}

func GenUUID() (string) {
 	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return ""
	}
 	// TODO: verify the two lines implement RFC 4122 correctly
	uuid[8] = (uuid[8] | 0x80) & 0xBf // variant bits see page 5
	uuid[4] = (uuid[6] | 0x40) & 0x4f // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid)
}

func (u *UUID) String() string {
        return fmt.Sprintf("%x-%x-%x-%x-%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
}

