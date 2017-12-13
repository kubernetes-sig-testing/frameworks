package test_test

import (
	"fmt"
	"io"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	. "k8s.io/kubectl/pkg/framework/test"
	"k8s.io/kubectl/pkg/framework/test/testfakes"
)

var _ = Describe("Etcd", func() {
	var (
		fakeSession        *testfakes.FakeSimpleSession
		fakeDataDirManager *testfakes.FakeDataDirManager
		fakePathFinder     *testfakes.FakeBinPathFinder
		fakeAddressManager *testfakes.FakeAddressManager
		etcd               *Etcd
	)

	BeforeEach(func() {
		fakeSession = &testfakes.FakeSimpleSession{}
		fakeDataDirManager = &testfakes.FakeDataDirManager{}
		fakePathFinder = &testfakes.FakeBinPathFinder{}
		fakeAddressManager = &testfakes.FakeAddressManager{}

		etcd = &Etcd{
			AddressManager: fakeAddressManager,
			PathFinder:     fakePathFinder.Spy,
			DataDirManager: fakeDataDirManager,
		}
	})

	Describe("starting and stopping etcd", func() {
		Context("when given a path to a binary that runs for a long time", func() {
			It("can start and stop that binary", func() {
				sessionBuffer := gbytes.NewBuffer()
				fmt.Fprintf(sessionBuffer, "Everything is dandy")
				fakeSession.BufferReturns(sessionBuffer)

				fakeSession.ExitCodeReturnsOnCall(0, -1)
				fakeSession.ExitCodeReturnsOnCall(1, 143)

				fakePathFinder.ReturnsOnCall(0, "/path/to/some/etcd")
				fakeAddressManager.InitializeReturns(1234, "this.is.etcd.listening.for.clients", nil)

				etcd.ProcessStarter = func(command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					Expect(command.Args).To(ContainElement(fmt.Sprintf("--advertise-client-urls=http://%s:%d", "this.is.etcd.listening.for.clients", 1234)))
					Expect(command.Args).To(ContainElement(fmt.Sprintf("--listen-client-urls=http://%s:%d", "this.is.etcd.listening.for.clients", 1234)))
					Expect(command.Path).To(Equal("/path/to/some/etcd"))
					fmt.Fprint(err, "serving insecure client requests on this.is.etcd.listening.for.clients:1234")
					return fakeSession, nil
				}

				By("Starting the Etcd Server")
				err := etcd.Start()
				Expect(err).NotTo(HaveOccurred())

				By("...in turn calling the PathFinder")
				Expect(fakePathFinder.CallCount()).To(Equal(1))
				Expect(fakePathFinder.ArgsForCall(0)).To(Equal("etcd"))

				By("...in turn calling using the AddressManager")
				Expect(fakeAddressManager.InitializeCallCount()).To(Equal(1))
				Expect(fakeAddressManager.InitializeArgsForCall(0)).To(Equal("localhost"))

				By("...in turn using the DataDirManager")
				Expect(fakeDataDirManager.CreateCallCount()).To(Equal(1))

				Eventually(etcd).Should(gbytes.Say("Everything is dandy"))
				Expect(fakeSession.ExitCodeCallCount()).To(Equal(0))
				Expect(etcd).NotTo(gexec.Exit())
				Expect(fakeSession.ExitCodeCallCount()).To(Equal(1))
				Expect(fakeDataDirManager.CreateCallCount()).To(Equal(1))

				By("Stopping the Etcd Server")
				etcd.Stop()

				Expect(fakeDataDirManager.DestroyCallCount()).To(Equal(1))
				Expect(etcd).To(gexec.Exit(143))
				Expect(fakeSession.TerminateCallCount()).To(Equal(1))
				Expect(fakeSession.WaitCallCount()).To(Equal(1))
				Expect(fakeSession.ExitCodeCallCount()).To(Equal(2))
				Expect(fakeDataDirManager.DestroyCallCount()).To(Equal(1))
			})
		})

		Context("when the data directory cannot be created", func() {
			It("propagates the error", func() {
				fakeDataDirManager.CreateReturnsOnCall(0, "", fmt.Errorf("Error on directory creation."))

				etcd.ProcessStarter = func(Command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					Expect(true).To(BeFalse(),
						"the etcd process starter shouldn't be called if getting a free port fails")
					return nil, nil
				}

				err := etcd.Start()
				Expect(err).To(MatchError(ContainSubstring("Error on directory creation.")))
			})
		})

		Context("when the address manager fails to get a new address", func() {
			It("propagates the error and does not start any process", func() {
				fakeAddressManager.InitializeReturns(0, "", fmt.Errorf("some error finding a free port"))

				etcd.ProcessStarter = func(Command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					Expect(true).To(BeFalse(),
						"the etcd process starter shouldn't be called if getting a free port fails")
					return nil, nil
				}

				Expect(etcd.Start()).To(MatchError(ContainSubstring("some error finding a free port")))
			})
		})

		Context("when  the starter returns an error", func() {
			It("propagates the error", func() {
				etcd.ProcessStarter = func(command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					return nil, fmt.Errorf("Some error in the starter.")
				}

				err := etcd.Start()
				Expect(err).To(MatchError(ContainSubstring("Some error in the starter.")))
			})
		})

		Context("when we try to stop a server that hasn't been started", func() {
			It("is a noop and does not call exit on the session", func() {
				etcd.ProcessStarter = func(command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					return fakeSession, nil
				}
				etcd.Stop()
				Expect(fakeSession.ExitCodeCallCount()).To(Equal(0))
			})
		})
	})

	Describe("querying the server for its URL", func() {
		It("can be queried for the URL it listens on", func() {
			fakeAddressManager.HostReturns("the.host.for.etcd", nil)
			fakeAddressManager.PortReturns(6789, nil)
			apiServerURL, err := etcd.URL()
			Expect(err).NotTo(HaveOccurred())
			Expect(apiServerURL).To(Equal("http://the.host.for.etcd:6789"))
		})

		Context("when we query for the URL before starting the server", func() {
			Context("and so the addressmanager fails to give us a port", func() {
				It("propagates the failure", func() {
					fakeAddressManager.PortReturns(0, fmt.Errorf("zort"))
					_, err := etcd.URL()
					Expect(err).To(MatchError(ContainSubstring("zort")))
				})
			})
			Context("and so the addressmanager fails to give us a host", func() {
				It("propagates the failure", func() {
					fakeAddressManager.HostReturns("", fmt.Errorf("bam!"))
					_, err := etcd.URL()
					Expect(err).To(MatchError(ContainSubstring("bam!")))
				})
			})
		})
	})
})
