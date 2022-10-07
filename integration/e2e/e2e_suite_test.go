/*

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-evm/integration/helpers"
	"github.com/hyperledger/fabric/integration/nwo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEndToEnd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EndToEnd Suite")
}

var components *nwo.Components

var _ = SynchronizedBeforeSuite(func() []byte {
	components = helpers.Build()

	payload, err := json.Marshal(components)
	Expect(err).NotTo(HaveOccurred())

	return payload
}, func(payload []byte) {
	err := json.Unmarshal(payload, &components)
	Expect(err).NotTo(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	components.Cleanup()
})
