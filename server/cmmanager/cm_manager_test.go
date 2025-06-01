package cmmanager

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

// This test file should represent what the client sends over the network with different test cases
func TestTypeParsing(t *testing.T) {
	// convert this into table driven test
	var data []byte = make([]byte, 2)

	binary.BigEndian.PutUint16(data, UPLOAD_CMD)
	buff := bytes.NewReader(data)

	cmd, err := handleTypeParsing(buff)
	if err != nil {
		t.Errorf("expected nil err, got %q instead", err)
	}
	if cmd != UPLOAD_CMD {
		t.Errorf("expected %d cmd, got %d instead", UPLOAD_CMD, cmd)
	}
}

func TestHandleFileNameParsing(t *testing.T) {
	// convert this into table driven test
	const fileName = "nTk25YFqGbdX[!zS@@]AkJ!P7@wML!QK5qUnif.rf;8GqLRV;Xr$jr*wZ{ZP94yS"
	buff := strings.NewReader(fileName)

	parsedFileName, err := handleFileNameParsing(buff, 64)
	if err != nil {
		t.Fatalf("expected nil err, got %q instead", err)
	}
	fmt.Println(fileName, parsedFileName)
	if parsedFileName != fileName {
		t.Errorf("expected %s cmd, got %s instead", fileName, parsedFileName)
	}
}
