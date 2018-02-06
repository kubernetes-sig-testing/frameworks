package integration_tests

import (
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kubernetes-sig-testing/frameworks/integration"
)

var _ = Describe("APIServer", func() {
	Context("when calling Config()", func() {
		It("returns a valid rest.Config", func() {
			etcd := &Etcd{}
			Expect(etcd.Start()).To(Succeed())
			defer func() {
				Expect(etcd.Stop()).To(Succeed())
			}()

			apiServer := &APIServer{EtcdURL: etcd.URL}
			Expect(apiServer.Start()).To(Succeed())
			defer func() {
				Expect(apiServer.Stop()).To(Succeed())
			}()

			c, err := apiServer.Config()
			Expect(err).NotTo(HaveOccurred())
			Expect(c.Host).To(Equal(apiServer.URL.String()))
		})

		Context("and the APIServer has not been started yet", func() {
			It("propagates the error", func() {
				apiServer := &APIServer{
					EtcdURL: &url.URL{Scheme: "http", Host: "localhost:0"},
				}
				_, err := apiServer.Config()
				Expect(err).To(MatchError(
					ContainSubstring("expected URL to be configured"),
				))
			})
		})
	})
})
