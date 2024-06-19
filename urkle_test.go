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
		name    string
		data    map[string]string
		outcome int
	}{
		// {
		// 	name:    "one",
		// 	data:    map[string]string{"a": "0000"},
		// 	outcome: 1,
		// },
		// {
		// 	name:    "two",
		// 	data:    map[string]string{"a": "0000", "b": "1100"},
		// 	outcome: 1,
		// },
		// {
		// 	name:    "three",
		// 	data:    map[string]string{"a": "0000", "b": "1100", "c": "1101"},
		// 	outcome: 1,
		// },
		// {
		// 	name:    "four",
		// 	data:    map[string]string{"a": "0000", "b": "1100", "c": "1101", "d": "1000"},
		// 	outcome: 1,
		// },

		{
			name:    "example",
			data:    map[string]string{"a": string([]byte{0x0, 0x0, 0x0, 0x0}), "b": string([]byte{0x1, 0x1, 0x0, 0x0}), "c": string([]byte{0x1, 0x1, 0x0, 0x1}), "d": string([]byte{0x1, 0x0, 0x0, 0x0}), "e": string([]byte{0x1, 0x0, 0x0, 0x1})},
			outcome: 1,
		},

		// {
		// 	name:    "example",
		// 	data:    map[string]string{"a": "00000", "b": "11000", "c": "11011", "d": "10000", "e": "10001", "f": "11110"},
		// 	outcome: 1,
		// },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			urkel := NewBinaryUrkelTrie()
			for v, b := range tc.data {
				urkel.Insert(b, v)
				//urkel = AddToUrkle(CreateLeafithKeyValue(b, b), urkel, 0, 0)
			}
			fmt.Printf("\n%v\n", urkel)
			assert.Equal(t, tc.outcome, 1)
		})
	}

}
