package tools

import(
	"testing"
	"fmt"
	"bytes"
	"encoding/hex"
	"strings"
)

func TestBuildServerRequest(t *testing.T){

	const pk = "\xea\x9b\xebU\x97G+%\x9cdp\x89\x1cn\x05\xad\x81\xb8\xc6\xa8Ny\x9d\xe79\xf41\x03\\\xbaz\x1f"
	const sk = "\xa5]\xb2\xc44\x19d,E\xb4Q\xb6Nx\xba\x82\xadl\x8e\xf3>Y>z\\\xba7{a\x07\x14\x7f"

	bPk, _ := hexArToByteAr(pk)
	bSk, _ := hexArToByteAr(sk)

	expectedResult := make(map[string][]byte)
	expectedResult["Secret"] = []byte("rM1WudxfYch3dLuvqAvgIPMPR18+tM6VNFxivyr+obXaFL4n3VDE70RKPzm5ILqhtMWwZpMrkSizPIpjqiIRhTa7XF9zm/BoqaigbcgRsJM=")
	expectedResult["PublicKey"] = []byte("6pvrVZdHKyWcZHCJHG4FrYG4xqhOeZ3nOfQxA1y6eh8=")
	expectedResult["ServerKey"] = []byte("MTExMTExMTExMTExMTExMTExMTExMTExMTExMTExMTE=")
	expectedResult["Nonce"] = []byte("MogS0NAB3wYdTpOuhlQP0sFnszw1GlGq")

	authenticator := Authenticator{}
	fmt.Println(bSk)
	data := authenticator.BuildServerRequest(bPk, bSk, "secret", []byte(strings.Repeat("1", 32)))

	fmt.Println(data)
	fmt.Println(expectedResult)

	if !bytes.Equal(data["Secret"], expectedResult["Secret"]) {
		t.Fail()
	}
	if !bytes.Equal(data["PublicKey"], expectedResult["PublicKey"]) {
		t.Fail()
	}
	if !bytes.Equal(data["Nonce"], expectedResult["Nonce"]) {
		t.Fail()
	}
	if !bytes.Equal(data["ServerKey"], expectedResult["ServerKey"]) {
		t.Fail()
	}
}

func hexArToByteAr(str string) ([]byte, error){
	var s string
	for i := 0; i < len(str); i++ {
		s += fmt.Sprintf("%x", str[i])
	}

	decoded, err := hex.DecodeString(s)

	return decoded, err
}