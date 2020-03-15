// +build libfuzzer

package blocks

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/go-ssz"
	stateTrie "github.com/prysmaticlabs/prysm/beacon-chain/state"
	p2ppb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
)

const PanicOnError = "true"

type InputBlockHeader struct {
	StateID uint16
	Block *ethpb.BeaconBlock
}

func bazelFileBytes(path string) ([]byte, error) {
	filepath, err := bazel.Runfile(path)
	if err != nil {
		return nil, err
	}
	fileBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}

func Fuzz(b []byte) []byte {
	input := &InputBlockHeader{}
	if err := ssz.Unmarshal(b, input); err != nil {
		return fail(err)
	}
	b, err := bazelFileBytes(fmt.Sprintf("external/sigp_beacon_fuzz_corpora/0-9-4/mainnet/beaconstate/%d", input.StateID))
	if err != nil {
		return fail(err)
	}
	s := &p2ppb.BeaconState{}
	if err := ssz.Unmarshal(b, s); err != nil {
		return fail(err)
	}
	st, err := stateTrie.InitializeFromProto(s)
	if err != nil {
		return fail(err)
	}
	post, err := ProcessBlockHeaderNoVerify(st, input.Block)
	if err != nil {
		return fail(err)
	}

	result, err := ssz.Marshal(post.InnerStateUnsafe())
	if err != nil {
		return fail(err)
	}
	return result
}

func fail(err error) []byte {
	if strings.ToUpper(PanicOnError) == "TRUE" {
		panic(err)
	}
	return nil
}
