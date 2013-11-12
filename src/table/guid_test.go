/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/11/13
 * Time: 10:23 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"strings"
	"testing"
)

func TestGuidToString(t *testing.T) {
	guid := GenUUID()
	if strings.Contains(guid, "-") {
		t.Error("GUID contains dashes.")
	}
}

