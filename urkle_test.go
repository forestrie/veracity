package veracity

import (
	"fmt"
	"os"
	"testing"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/stretchr/testify/assert"
)

// func TestCommonPrefix(t *testing.T) {

// 	logger.New("TestVerifyList")
// 	defer logger.OnExit()
// 	url := os.Getenv("TEST_INTEGRATION_FORESTRIE_BLOBSTORE_URL")
// 	logger.Sugar.Infof("url: '%s'", url)

// 	tests := []struct {
// 		name    string
// 		node    []byte
// 		value   []byte
// 		outcome int
// 	}{
// 		{
// 			name:    "same",
// 			node:    []byte{0x00, 0x01, 0x02, 0x00},
// 			value:   []byte{0x00, 0x01, 0x02, 0x00},
// 			outcome: 32,
// 		},
// 		{
// 			name:    "simple offset",
// 			node:    []byte{0x00, 0x01, 0x02, 0x00},
// 			value:   []byte{0x00, 0x01, 0x02, 0x01},
// 			outcome: 31,
// 		},
// 		{
// 			name:    "is on the right",
// 			node:    []byte{0x00, 0x01, 0x02, 0x02},
// 			value:   []byte{0x00, 0x01, 0x02, 0x01},
// 			outcome: 30,
// 		},
// 		{
// 			name:    "has gap",
// 			node:    []byte{0x00, 0x01, 0x02, 0x00},
// 			value:   []byte{0x00, 0x01, 0x02, 0x02},
// 			outcome: 30,
// 		},
// 		{
// 			name:    "wrap around",
// 			node:    []byte{0x00, 0x01, 0x02, 0xff},
// 			value:   []byte{0xff, 0x01, 0x03, 0x00},
// 			outcome: 0,
// 		},
// 		{
// 			name:    "wrap around",
// 			node:    []byte{0x01, 0x01, 0x02, 0xff},
// 			value:   []byte{0x05, 0x01, 0x03, 0x00},
// 			outcome: 0,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			r := commonPrefix(tc.node, tc.value)
// 			assert.Equal(t, tc.outcome, r)
// 			fmt.Printf("vvvv %v\n", tc.value)
// 		})
// 	}

// }

// func TestCommonDepth(t *testing.T) {

// 	logger.New("TestVerifyList")
// 	defer logger.OnExit()
// 	url := os.Getenv("TEST_INTEGRATION_FORESTRIE_BLOBSTORE_URL")
// 	logger.Sugar.Infof("url: '%s'", url)

// 	tests := []struct {
// 		name    string
// 		value   int
// 		length  int
// 		outcome int
// 	}{
// 		{
// 			name:    "same",
// 			value:   5,
// 			length:  4,
// 			outcome: 88,
// 		},
// 		{
// 			name:    "simple offset",
// 			value:   31,
// 			length:  32,
// 			outcome: 5,
// 		},
// 		{
// 			name:    "is on the right",
// 			value:   30,
// 			length:  32,
// 			outcome: 5,
// 		},
// 		{
// 			name:    "has gap",
// 			value:   5,
// 			length:  32,
// 			outcome: 3,
// 		},
// 		{
// 			name:    "wrap around",
// 			value:   255,
// 			length:  256,
// 			outcome: 8,
// 		},
// 		{
// 			name:    "wrap around",
// 			value:   256,
// 			length:  256,
// 			outcome: 9,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			r := commonDepth(tc.value, tc.length)
// 			assert.Equal(t, tc.outcome, r)
// 		})
// 	}

// }
func TestBit(t *testing.T) {

	logger.New("TestVerifyList")
	defer logger.OnExit()
	url := os.Getenv("TEST_INTEGRATION_FORESTRIE_BLOBSTORE_URL")
	logger.Sugar.Infof("url: '%s'", url)

	tests := []struct {
		name    string
		value   []byte
		pos     int
		outcome bool
	}{
		{
			name:    "simple",
			value:   []byte{0b10},
			pos:     1,
			outcome: true,
		},
		{
			name:    "simple",
			value:   []byte{0b10, 0b11},
			pos:     8,
			outcome: true,
		},
		{
			name:    "hit on first",
			value:   []byte{0b1},
			pos:     0,
			outcome: true,
		},
		{
			name:    "another hit",
			value:   []byte{0b11111111},
			pos:     7,
			outcome: true,
		},
		{
			name:    "miss",
			value:   []byte{0b1000000},
			pos:     7,
			outcome: false,
		},
		{
			name:    "miss 2",
			value:   []byte{0b1000000},
			pos:     5,
			outcome: false,
		},
		{
			name:    "hit 2",
			value:   []byte{0b1000000},
			pos:     6,
			outcome: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := bit(tc.pos, tc.value)
			assert.Equal(t, tc.outcome, r)
		})
	}

}

func TestAddToUrkle(t *testing.T) {

	logger.New("TestVerifyList")
	defer logger.OnExit()
	url := os.Getenv("TEST_INTEGRATION_FORESTRIE_BLOBSTORE_URL")
	logger.Sugar.Infof("url: '%s'", url)

	tests := []struct {
		name          string
		data          map[string][]byte
		find          []byte
		found         bool
		expectedPath  []string
		expectedValue string
	}{
		// slightly bigger tree with a 4 byte keys
		{
			name: "big",
			data: map[string][]byte{
				"a": {0x0, 0x0, 0x0, 0x0},
				"b": {0x1, 0x1, 0x0, 0x0},
				"c": {0x1, 0x1, 0x0, 0x1},
				"d": {0x1, 0x0, 0x0, 0x0},
				"e": {0x1, 0x0, 0x0, 0x1},
				"f": {0xff, 0x0, 0x0, 0x1}},
			find:  []byte{0x1, 0x0, 0x0, 0x0},
			found: true,
			expectedPath: []string{
				"ea328a7c32777d9588adf254cea119d807b426a6564e286da8bbff891d7c9299",
				"b07a29e62a2ed9520d74203e412780a0f6058b951c12a511272b3341e6e2ab34",
				"8a27a67484bb2a1e3ede97d52634f5e9f0300eb2ac1d2fb44fb91463443b52be",
				"b79610444fc284e694c006a22572f9c6787b63050a9aef731fd7b8fdf92cd9cf",
				"436a65b0eda87a7f0a5843667dfe4a8b98d6af9e7cbdc1ec453a97a0af62c193",
				"e3c6742990e888b490cf09b9ec3e3419e37a745b4b512274cf36fe2d6dd33a78",
				"ef25d0943286c56f562fc4e2b3ba3238587ec4856a54b819c33ec7446716a92a",
				"24dfd3bd2a5057ed26f56b74049630c3ed446d0998241dbdabfc4a65f108092e",
				"4b6ca2f50c4303b39cd1cb71137ec291faed42119c49613666b85148acb228b9",
				"66ffca459a114b860958f446c6c08f775017005b47d3f6144431221b3b2ac898",
				"fc5819d78740e6cb495ecdb8010927266bdf1e6a1ceed1c62d2fbf1e39ff1f4c",
				"0a9496a10280061d11b946c736d1fb74348b9367e572929cee03577a6d795290",
				"210e8b144ff8f9bbcff8ee239f04acdef38ed14d969c323959b288d00ab5016f",
				"bacbc96ea0582ca779c6353cfd540f27d7371aaae162b70c94758f30b8ef695d",
				"fd7ac75a17282f2c8eb4df80e7f29150ebaaf0c39416aa8e0d04ad09ecae73f9",
				"178ff4e1d91f46da790299bf853eff7fa958ad0c54c5dbf50e7fdd8f9eae8e5c",
				"567ee242c79ee88afadd2d2ef7615cb67e6cb4fb5650b0a9a57362cac3c86104",
				"92b1c319b011c1b1d8bdd22086f3cd10dc2ee0acb7869e4856516f783d8a8acf",
				"65c20bf647915a7f8d77caa3eabeaf8adc686609f81199a719cf10a040afd9d8",
				"9467a533af393497f7dcf7ab4dac9d85436c38f7e19c7a93b65b306f782ce762",
				"8653b7c5265da397906ec7f50a032a42a1b32b69a0fec8c48f5d9d2753f399c4",
				"d8e91c8a62d60f289363fcd0b43d21dfe3ad4123c59bb67191e9a73ed712936f",
				"ea12f03733ae572c944770b3ba7c6c53c79cde809d1612d40f16b5b4c29ca51f",
				"4bd829cb47368d619620fd881fe283f631bb0fe58633bc27390068b114941af9",
				"160dea445485eb5969771dbe7e70eaa70e1291b213f42d7547306ab12d682cc6",
				"db6312be0cf8ac1b634610bc1f0e1b06609b6cc3aacefa13cfadcff31d8c02fb",
				"f9fc26ecbc9c376d754af149aac41befa4d9f3f3867de92348a759400e1a97fa",
				"b96900008f9adc9074b02d5de3acd1953cac6fa3b37e9a7a72dabd41db124a61",
				"911859c84aba9720d06963090343494979cc7796050891375487e78026e4cdd4",
				"2a8104f160a51dacf62d9e8f80c48c6899a65364ae1ea7e81c2021f810d35d54",
				"ec44b4b1ee32431874f5ae6d8f496a9e486bfe232d2656e6ed4b5463b320692e",
				"00edc6dcc7db1e573cb910e42d53d26542ee6a7e7174f0f76b917c977931f3e5",
				"18ac3e7343f016890c510e93f935261169d9e3f565436429830faf0934f4f8e4"},
			expectedValue: "d",
		},
		// this is the tree from examples@ https://handshake.org/files/handshake.txt
		// 		   R
		// 	     /   \
		// 	    /     \
		// 	   a      /\
		// 		     /  \
		// 		    /    \
		// 		   /      \
		// 	      /        \
		// 	     /\        /\
		// 	    /  \      /  \
		// 	   /\   x    /\   x
		//    /  \      /  \
		//   d    e    b    c
		//
		// all involved hashes:
		// a - ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb
		// b - 3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d
		// c - 2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6
		// d - 18ac3e7343f016890c510e93f935261169d9e3f565436429830faf0934f4f8e4
		// e - 3f79bb7b435b05321651daefd374cdc681dc06faa65e374e38337b88ca046dea

		// x - 0000000000000000000000000000000000000000000000000000000000000000

		// de      - 00edc6dcc7db1e573cb910e42d53d26542ee6a7e7174f0f76b917c977931f3e5
		// bc      - 6ab3ff83fd75c25215d659c9b61e2e8c83913b7c600e7f0a74303e1aeb61d515
		// dex     - ec44b4b1ee32431874f5ae6d8f496a9e486bfe232d2656e6ed4b5463b320692e
		// bcx     - 1de8679edf9aa48e2de9b221439f63ebb6a26977cc467ad6bb653bf5b7906ef4
		// dexbcx  - d422b34896d1ba6138fc2f704874d2123f723df5f77fdc29b87f6133b4087729
		// adexbcx - f76c803e6995719137a9d534ba508b5bb0ef4a32e76cddac702eab3c3ea62e27
		{
			name: "example c",
			data: map[string][]byte{
				"a": {0b0},
				"b": {0b11000000},
				"c": {0b11010000},
				"d": {0b10000000},
				"e": {0b10010000}},
			find:  []byte{0b11010000},
			found: true,
			expectedPath: []string{
				"f76c803e6995719137a9d534ba508b5bb0ef4a32e76cddac702eab3c3ea62e27",
				"d422b34896d1ba6138fc2f704874d2123f723df5f77fdc29b87f6133b4087729",
				"1de8679edf9aa48e2de9b221439f63ebb6a26977cc467ad6bb653bf5b7906ef4",
				"6ab3ff83fd75c25215d659c9b61e2e8c83913b7c600e7f0a74303e1aeb61d515",
				"2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6"},
			expectedValue: "c",
		},
		{
			name: "example a",
			data: map[string][]byte{
				"a": {0b0},
				"b": {0b11000000},
				"c": {0b11010000},
				"d": {0b10000000},
				"e": {0b10010000}},
			find:  []byte{0b0},
			found: true,
			expectedPath: []string{
				"f76c803e6995719137a9d534ba508b5bb0ef4a32e76cddac702eab3c3ea62e27",
				"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb"},
			expectedValue: "a",
		},
		{
			name: "example e",
			data: map[string][]byte{
				"a": {0b0},
				"b": {0b11000000},
				"c": {0b11010000},
				"d": {0b10000000},
				"e": {0b10010000}},
			find:  []byte{0b10010000},
			found: true,
			expectedPath: []string{
				"f76c803e6995719137a9d534ba508b5bb0ef4a32e76cddac702eab3c3ea62e27",
				"d422b34896d1ba6138fc2f704874d2123f723df5f77fdc29b87f6133b4087729",
				"ec44b4b1ee32431874f5ae6d8f496a9e486bfe232d2656e6ed4b5463b320692e",
				"00edc6dcc7db1e573cb910e42d53d26542ee6a7e7174f0f76b917c977931f3e5",
				"3f79bb7b435b05321651daefd374cdc681dc06faa65e374e38337b88ca046dea"},
			expectedValue: "e",
		},

		{
			name: "example missing",
			data: map[string][]byte{
				"a": {0b0},
				"b": {0b11000000},
				"c": {0b11010000},
				"d": {0b10000000},
				"e": {0b10010000}},
			find:  []byte{0b10010010},
			found: false,
			expectedPath: []string{
				"f76c803e6995719137a9d534ba508b5bb0ef4a32e76cddac702eab3c3ea62e27",
				"d422b34896d1ba6138fc2f704874d2123f723df5f77fdc29b87f6133b4087729",
				"ec44b4b1ee32431874f5ae6d8f496a9e486bfe232d2656e6ed4b5463b320692e",
				"00edc6dcc7db1e573cb910e42d53d26542ee6a7e7174f0f76b917c977931f3e5",
				"3f79bb7b435b05321651daefd374cdc681dc06faa65e374e38337b88ca046dea"},
			expectedValue: "e",
		},
		{
			name: "example missing terminating",
			data: map[string][]byte{
				"a": {0b0},
				"b": {0b11000000},
				"c": {0b11010000},
				"d": {0b10000000},
				"e": {0b10010000}},
			find:  []byte{0b10110000},
			found: false,
			expectedPath: []string{
				"f76c803e6995719137a9d534ba508b5bb0ef4a32e76cddac702eab3c3ea62e27",
				"d422b34896d1ba6138fc2f704874d2123f723df5f77fdc29b87f6133b4087729",
				"ec44b4b1ee32431874f5ae6d8f496a9e486bfe232d2656e6ed4b5463b320692e",
				"0000000000000000000000000000000000000000000000000000000000000000"},
			expectedValue: "",
		},
		{
			name: "example missing left",
			data: map[string][]byte{
				"a": {0b0},
				"b": {0b11000000},
				"c": {0b11010000},
				"d": {0b10000000},
				"e": {0b10010000}},
			find:  []byte{0b00000001},
			found: false,
			expectedPath: []string{
				"f76c803e6995719137a9d534ba508b5bb0ef4a32e76cddac702eab3c3ea62e27",
				"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb"},
			expectedValue: "a",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			urkel := NewBinaryUrkelTrie()
			for v, b := range tc.data {
				urkel.Insert(b, v)
				//urkel = AddToUrkle(CreateLeafithKeyValue(b, b), urkel, 0, 0)
			}
			fmt.Printf("\n%v\n", urkel)

			path, val, found := urkel.GetPath(tc.find)

			assert.Equal(t, tc.found, found)
			assert.Equal(t, tc.expectedValue, val)
			assert.ElementsMatch(t, tc.expectedPath, path)
		})
	}
}
