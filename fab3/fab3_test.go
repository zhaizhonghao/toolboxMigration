/*

SPDX-License-Identifier: Apache-2.0
*/

package fab3_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/hyperledger/fabric-chaincode-evm/fab3"
	"github.com/hyperledger/fabric-chaincode-evm/fab3/mocks"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fab3", func() {

	var (
		proxy          *fab3.Fab3
		proxyAddr      string
		mockEthService *mocks.MockEthService
		req            *http.Request
		proxyDoneChan  chan struct{}
		client         *http.Client
		port           int
	)

	BeforeEach(func() {
		port = config.GinkgoConfig.ParallelNode + 5000
		mockEthService = &mocks.MockEthService{}
		client = &http.Client{}

		proxyDoneChan = make(chan struct{}, 1)
		var err error
		proxy = fab3.NewFab3(mockEthService, port)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when fab3 successfully starts", func() {
		BeforeEach(func() {
			go func(proxy *fab3.Fab3, proxyDoneChan chan struct{}) {
				proxy.Start()

				// Close proxy done chan to signify proxy has exited
				close(proxyDoneChan)
			}(proxy, proxyDoneChan)

			Eventually(proxyDoneChan).ShouldNot(Receive())

			proxyAddr = fmt.Sprintf("http://localhost:%d", port)

			//Ensure the server is up before starting the test
			Eventually(func() error {
				conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
				if err == nil {
					defer conn.Close()
				}
				return err
			}).Should(Succeed())

			mockEthService.GetCodeStub = func(r *http.Request, arg *string, reply *string) error {
				*reply = "0x11110"
				return nil
			}

			//curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getCode","params":["0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b", "0x2"],"id":1}'
			body := strings.NewReader(`{"jsonrpc":"2.0","method":"eth_getCode","params":["0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b"],"id":1}`)

			var err error
			req, err = http.NewRequest("POST", proxyAddr, body)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
		})

		AfterEach(func() {
			err := proxy.Shutdown()
			Expect(err).ToNot(HaveOccurred())

			Eventually(proxyDoneChan).Should(BeClosed())
		})

		It("starts a server that uses the provided ethservice", func() {
			resp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())

			type responseBody struct {
				JsonRPC string `json:"jsonrpc"`
				ID      int    `json:"id"`
				Result  string `json:"result"`
			}
			expectedBody := responseBody{JsonRPC: "2.0", ID: 1, Result: "0x11110"}

			rBody, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			var respBody responseBody
			err = json.Unmarshal(rBody, &respBody)
			Expect(err).ToNot(HaveOccurred())

			Expect(respBody).To(Equal(expectedBody))
		})

		It("starts a server that uses the hardcoded netservice", func() {
			var err error
			body := strings.NewReader(`{"jsonrpc":"2.0","method":"net_version","id":1}`)
			req, err = http.NewRequest("POST", proxyAddr, body)
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())

			type responseBody struct {
				JsonRPC string `json:"jsonrpc"`
				ID      int    `json:"id"`
				Result  string `json:"result"`
			}
			expectedBody := responseBody{JsonRPC: "2.0", ID: 1, Result: hex.EncodeToString([]byte(fab3.NetworkID))}

			rBody, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			var respBody responseBody
			err = json.Unmarshal(rBody, &respBody)
			Expect(err).ToNot(HaveOccurred())

			Expect(respBody).To(Equal(expectedBody))
		})

		Context("when the request has Cross-Origin Resource Sharing Headers", func() {
			BeforeEach(func() {
				var err error
				body := strings.NewReader("")
				//OPTIONS pre-check used to see the CORS options of the server
				req, err = http.NewRequest("OPTIONS", proxyAddr, body)
				Expect(err).ToNot(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Origin", "[http://example.com]")
			})

			Context("when the request method is POST", func() {
				BeforeEach(func() {
					//curl -X OPTIONS http://localhost:5000 -H "Origin: http://example.com" -H "Access-Control-Request-Method: POST"
					req.Header.Set("Access-Control-Request-Method", "POST")
				})

				It("successfully processes the request", func() {
					client := &http.Client{}
					resp, err := client.Do(req)
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(resp.Header.Get("Access-Control-Allow-Origin")).To(Equal("*"))
				})
			})

			Context("when the request method is not POST", func() {
				BeforeEach(func() {
					//curl -X OPTIONS http://localhost:5000 -H "Origin: http://example.com" -H "Access-Control-Request-Method: GET"
					req.Header.Set("Access-Control-Request-Method", "GET")
				})

				It("should fail", func() {
					client := &http.Client{}
					resp, err := client.Do(req)
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
				})
			})
		})
	})

	Context("when the requested port is already bound", func() {
		var ln net.Listener

		BeforeEach(func() {
			var err error
			ln, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := ln.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error and does not start", func() {
			err := proxy.Start()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when Fab3 has no HTTP Server ", func() {
		It("does not error when being shutdown", func() {
			proxy.HTTPServer = nil
			err := proxy.Shutdown()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
